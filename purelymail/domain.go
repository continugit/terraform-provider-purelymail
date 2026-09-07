package purelymail

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type domainModel struct {
	Name                  types.String     `tfsdk:"name"`
	AllowAccountReset     types.Bool       `tfsdk:"allow_account_reset"`
	SymbolicSubaddressing types.Bool       `tfsdk:"symbolic_subaddressing"`
	IsShared              types.Bool       `tfsdk:"is_shared"`
	PassesMX              types.Bool       `tfsdk:"passes_mx"`
	PassesSPF             types.Bool       `tfsdk:"passes_spf"`
	PassesDKIM            types.Bool       `tfsdk:"passes_dkim"`
	PassesDMARC           types.Bool       `tfsdk:"passes_dmarc"`
	RecheckDNS            types.Bool       `tfsdk:"recheck_dns"`
	Timeouts              timeouts.Value   `tfsdk:"timeouts"`
	DNSRecords            []dnsRecordModel `tfsdk:"dns_records"`
}

type dnsRecordModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Value    types.String `tfsdk:"value"`
	Priority types.Int64  `tfsdk:"priority"`
}

func domainSchema(ctx context.Context) resourceschema.Schema {
	return resourceschema.Schema{Attributes: map[string]resourceschema.Attribute{
		"name":                   resourceschema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"allow_account_reset":    resourceschema.BoolAttribute{Optional: true, Computed: true},
		"symbolic_subaddressing": resourceschema.BoolAttribute{Optional: true, Computed: true},
		"is_shared":              resourceschema.BoolAttribute{Computed: true},
		"passes_mx":              resourceschema.BoolAttribute{Computed: true},
		"passes_spf":             resourceschema.BoolAttribute{Computed: true},
		"passes_dkim":            resourceschema.BoolAttribute{Computed: true},
		"passes_dmarc":           resourceschema.BoolAttribute{Computed: true},
		"recheck_dns":            resourceschema.BoolAttribute{Optional: true},
		"timeouts":               timeouts.Attributes(ctx, timeouts.Opts{Create: true}),
		"dns_records": resourceschema.ListNestedAttribute{
			Computed: true,
			NestedObject: resourceschema.NestedAttributeObject{Attributes: map[string]resourceschema.Attribute{
				"name":     resourceschema.StringAttribute{Computed: true},
				"type":     resourceschema.StringAttribute{Computed: true},
				"value":    resourceschema.StringAttribute{Computed: true},
				"priority": resourceschema.Int64Attribute{Computed: true},
			}},
		},
	}}
}

func setDomain(m *domainModel, d Domain) {
	m.Name = types.StringValue(d.Name)
	m.AllowAccountReset = types.BoolValue(d.AllowAccountReset)
	m.SymbolicSubaddressing = types.BoolValue(d.SymbolicSubaddressing)
	m.IsShared = types.BoolValue(d.IsShared)
	m.PassesMX = types.BoolValue(d.DNSSummary.PassesMX)
	m.PassesSPF = types.BoolValue(d.DNSSummary.PassesSPF)
	m.PassesDKIM = types.BoolValue(d.DNSSummary.PassesDKIM)
	m.PassesDMARC = types.BoolValue(d.DNSSummary.PassesDMARC)
}

type DomainResource struct{ client *Client }

func NewDomainResource() resource.Resource { return &DomainResource{} }
func (r *DomainResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "purelymail_domain"
}
func (r *DomainResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = domainSchema(ctx)
}
func (r *DomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client, _ = req.ProviderData.(*Client)
}
func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout, timeoutDiags := plan.Timeouts.Create(ctx, 10*time.Minute)
	resp.Diagnostics.Append(timeoutDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	createCtx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()
	found, err := waitForOwnershipTXT(createCtx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Check domain ownership DNS failed", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Ownership DNS record not found", "Timed out waiting for a TXT record containing purelymail_ownership_proof= at the domain apex. Create the record manually or with another Terraform provider, then retry.")
		return
	}
	var out Empty
	if err := r.client.Do(createCtx, "addDomain", map[string]string{"domainName": plan.Name.ValueString()}, &out); err != nil {
		resp.Diagnostics.AddError("Create domain failed", err.Error())
		return
	}
	found, err = r.waitForDomain(createCtx, plan.Name.ValueString(), &plan)
	if err != nil {
		resp.Diagnostics.AddError("Read domain after creation failed", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Domain was not added", "Purelymail acknowledged addDomain, but the domain is not visible in listDomains. Verify the ownership DNS record and complete the domain setup in Purelymail before retrying.")
		return
	}
	if err := r.setDNSRecords(createCtx, &plan); err != nil {
		resp.Diagnostics.AddError("Read domain DNS records failed", err.Error())
		return
	}
	resp.State.Set(ctx, &plan)
}

func waitForOwnershipTXT(ctx context.Context, domain string) (bool, error) {
	resolver := net.DefaultResolver
	for {
		records, err := resolver.LookupTXT(ctx, domain)
		if err == nil {
			for _, record := range records {
				if strings.Contains(record, "purelymail_ownership_proof=") {
					return true, nil
				}
			}
		}
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return false, nil
			}
			return false, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}
func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.read(ctx, state.Name.ValueString(), &state, &resp.Diagnostics) {
		resp.State.Set(ctx, &state)
	} else if !resp.Diagnostics.HasError() {
		resp.State.RemoveResource(ctx)
	}
}
func (r *DomainResource) read(ctx context.Context, name string, state *domainModel, diags *diag.Diagnostics) bool {
	var out listDomainsResponse
	if err := r.client.Do(ctx, "listDomains", map[string]bool{"includeShared": true}, &out); err != nil {
		diags.AddError("Read domain failed", err.Error())
		return false
	}
	for _, d := range out.Domains {
		if d.Name == name {
			setDomain(state, d)
			if err := r.setDNSRecords(ctx, state); err != nil {
				diags.AddError("Read domain DNS records failed", err.Error())
				return false
			}
			return true
		}
	}
	return false
}

func (r *DomainResource) setDNSRecords(ctx context.Context, state *domainModel) error {
	var ownership ownershipResponse
	if err := r.client.Do(ctx, "getOwnershipCode", map[string]any{}, &ownership); err != nil {
		return err
	}
	domain := state.Name.ValueString()
	state.DNSRecords = []dnsRecordModel{
		{Name: types.StringValue(domain), Type: types.StringValue("TXT"), Value: types.StringValue(ownership.Code), Priority: types.Int64Null()},
		{Name: types.StringValue(domain), Type: types.StringValue("MX"), Value: types.StringValue("mailserver.purelymail.com"), Priority: types.Int64Value(50)},
		{Name: types.StringValue(domain), Type: types.StringValue("TXT"), Value: types.StringValue("v=spf1 include:_spf.purelymail.com ~all"), Priority: types.Int64Null()},
		{Name: types.StringValue("purelymail1._domainkey." + domain), Type: types.StringValue("CNAME"), Value: types.StringValue("key1.dkimroot.purelymail.com"), Priority: types.Int64Null()},
		{Name: types.StringValue("purelymail2._domainkey." + domain), Type: types.StringValue("CNAME"), Value: types.StringValue("key2.dkimroot.purelymail.com"), Priority: types.Int64Null()},
		{Name: types.StringValue("purelymail3._domainkey." + domain), Type: types.StringValue("CNAME"), Value: types.StringValue("key3.dkimroot.purelymail.com"), Priority: types.Int64Null()},
		{Name: types.StringValue("_dmarc." + domain), Type: types.StringValue("CNAME"), Value: types.StringValue("dmarcroot.purelymail.com"), Priority: types.Int64Null()},
	}
	return nil
}

func (r *DomainResource) waitForDomain(ctx context.Context, name string, state *domainModel) (bool, error) {
	for attempt := 0; attempt < 10; attempt++ {
		var out listDomainsResponse
		if err := r.client.Do(ctx, "listDomains", map[string]bool{"includeShared": true}, &out); err != nil {
			return false, err
		}
		for _, domain := range out.Domains {
			if domain.Name == name {
				setDomain(state, domain)
				return true, nil
			}
		}
		if attempt < 9 {
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(time.Second):
			}
		}
	}
	return false, nil
}
func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan domainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"name": plan.Name.ValueString()}
	if !plan.AllowAccountReset.IsNull() {
		body["allowAccountReset"] = plan.AllowAccountReset.ValueBool()
	}
	if !plan.SymbolicSubaddressing.IsNull() {
		body["symbolicSubaddressing"] = plan.SymbolicSubaddressing.ValueBool()
	}
	if !plan.RecheckDNS.IsNull() {
		body["recheckDns"] = plan.RecheckDNS.ValueBool()
	}
	if err := r.client.Do(ctx, "updateDomainSettings", body, &Empty{}); err != nil {
		resp.Diagnostics.AddError("Update domain failed", err.Error())
		return
	}
	if r.read(ctx, plan.Name.ValueString(), &plan, &resp.Diagnostics) {
		resp.State.Set(ctx, &plan)
	} else if !resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Domain not found after update", "Purelymail did not return the updated domain from listDomains.")
	}
}
func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Do(ctx, "deleteDomain", map[string]string{"name": state.Name.ValueString()}, &Empty{}); err != nil {
		resp.Diagnostics.AddError("Delete domain failed", err.Error())
	}
}

type DomainsDataSource struct{ client *Client }
type domainsModel struct {
	IncludeShared types.Bool    `tfsdk:"include_shared"`
	Domains       []domainModel `tfsdk:"domains"`
}

func NewDomainsDataSource() datasource.DataSource { return &DomainsDataSource{} }
func (d *DomainsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "purelymail_domains"
}
func (d *DomainsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{Attributes: map[string]datasourceschema.Attribute{
		"include_shared": datasourceschema.BoolAttribute{Optional: true}, "domains": datasourceschema.ListNestedAttribute{Computed: true, NestedObject: datasourceschema.NestedAttributeObject{Attributes: domainDataAttrs()}},
	}}
}
func domainDataAttrs() map[string]datasourceschema.Attribute {
	return map[string]datasourceschema.Attribute{
		"name": datasourceschema.StringAttribute{Computed: true}, "allow_account_reset": datasourceschema.BoolAttribute{Computed: true}, "symbolic_subaddressing": datasourceschema.BoolAttribute{Computed: true}, "is_shared": datasourceschema.BoolAttribute{Computed: true}, "passes_mx": datasourceschema.BoolAttribute{Computed: true}, "passes_spf": datasourceschema.BoolAttribute{Computed: true}, "passes_dkim": datasourceschema.BoolAttribute{Computed: true}, "passes_dmarc": datasourceschema.BoolAttribute{Computed: true},
		"dns_records": datasourceschema.ListNestedAttribute{Computed: true, NestedObject: datasourceschema.NestedAttributeObject{Attributes: map[string]datasourceschema.Attribute{
			"name": datasourceschema.StringAttribute{Computed: true}, "type": datasourceschema.StringAttribute{Computed: true}, "value": datasourceschema.StringAttribute{Computed: true}, "priority": datasourceschema.Int64Attribute{Computed: true},
		}}},
	}
}
func (d *DomainsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client, _ = req.ProviderData.(*Client)
	}
}
func (d *DomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg domainsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out listDomainsResponse
	if err := d.client.Do(ctx, "listDomains", map[string]bool{"includeShared": cfg.IncludeShared.ValueBool()}, &out); err != nil {
		resp.Diagnostics.AddError("List domains failed", err.Error())
		return
	}
	cfg.Domains = make([]domainModel, len(out.Domains))
	for i, dom := range out.Domains {
		setDomain(&cfg.Domains[i], dom)
		if err := (&DomainResource{client: d.client}).setDNSRecords(ctx, &cfg.Domains[i]); err != nil {
			resp.Diagnostics.AddError("Read domain DNS records failed", err.Error())
			return
		}
	}
	resp.State.Set(ctx, &cfg)
}
