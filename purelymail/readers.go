package purelymail

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type domainLookup struct {
	Name                  types.String `tfsdk:"name"`
	AllowAccountReset     types.Bool   `tfsdk:"allow_account_reset"`
	SymbolicSubaddressing types.Bool   `tfsdk:"symbolic_subaddressing"`
	IsShared              types.Bool   `tfsdk:"is_shared"`
	PassesMX              types.Bool   `tfsdk:"passes_mx"`
	PassesSPF             types.Bool   `tfsdk:"passes_spf"`
	PassesDKIM            types.Bool   `tfsdk:"passes_dkim"`
	PassesDMARC           types.Bool   `tfsdk:"passes_dmarc"`
}

type DomainDataSource struct{ client *Client }

func NewDomainDataSource() datasource.DataSource { return &DomainDataSource{} }
func (d *DomainDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "purelymail_domain"
}
func (d *DomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{Attributes: map[string]datasourceschema.Attribute{
		"name":                   datasourceschema.StringAttribute{Required: true},
		"allow_account_reset":    datasourceschema.BoolAttribute{Computed: true},
		"symbolic_subaddressing": datasourceschema.BoolAttribute{Computed: true},
		"is_shared":              datasourceschema.BoolAttribute{Computed: true},
		"passes_mx":              datasourceschema.BoolAttribute{Computed: true},
		"passes_spf":             datasourceschema.BoolAttribute{Computed: true},
		"passes_dkim":            datasourceschema.BoolAttribute{Computed: true},
		"passes_dmarc":           datasourceschema.BoolAttribute{Computed: true},
	}}
}
func (d *DomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client, _ = req.ProviderData.(*Client)
	}
}
func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m domainLookup
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out listDomainsResponse
	if err := d.client.Do(ctx, "listDomains", map[string]bool{"includeShared": true}, &out); err != nil {
		resp.Diagnostics.AddError("List domains failed", err.Error())
		return
	}
	for _, dom := range out.Domains {
		if dom.Name == m.Name.ValueString() {
			m.AllowAccountReset = types.BoolValue(dom.AllowAccountReset)
			m.SymbolicSubaddressing = types.BoolValue(dom.SymbolicSubaddressing)
			m.IsShared = types.BoolValue(dom.IsShared)
			m.PassesMX = types.BoolValue(dom.DNSSummary.PassesMX)
			m.PassesSPF = types.BoolValue(dom.DNSSummary.PassesSPF)
			m.PassesDKIM = types.BoolValue(dom.DNSSummary.PassesDKIM)
			m.PassesDMARC = types.BoolValue(dom.DNSSummary.PassesDMARC)
			resp.State.Set(ctx, &m)
			return
		}
	}
	resp.Diagnostics.AddError("Domain not found", "The requested domain is not accessible to this account.")
}

type OwnershipCodeDataSource struct{ client *Client }
type ownershipModel struct {
	Code types.String `tfsdk:"code"`
}

func NewOwnershipCodeDataSource() datasource.DataSource { return &OwnershipCodeDataSource{} }
func (d *OwnershipCodeDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "purelymail_ownership_code"
}
func (d *OwnershipCodeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{Attributes: map[string]datasourceschema.Attribute{
		"code": datasourceschema.StringAttribute{
			Computed:    true,
			Description: "Complete TXT record value, including the purelymail_ownership_proof= prefix.",
		},
	}}
}
func (d *OwnershipCodeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client, _ = req.ProviderData.(*Client)
	}
}
func (d *OwnershipCodeDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var out ownershipResponse
	if err := d.client.Do(ctx, "getOwnershipCode", map[string]any{}, &out); err != nil {
		resp.Diagnostics.AddError("Get ownership code failed", err.Error())
		return
	}
	resp.State.Set(ctx, &ownershipModel{Code: types.StringValue(out.Code)})
}

type AccountCreditDataSource struct{ client *Client }
type creditModel struct {
	Credit types.String `tfsdk:"credit"`
}

func NewAccountCreditDataSource() datasource.DataSource { return &AccountCreditDataSource{} }
func (d *AccountCreditDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "purelymail_account_credit"
}
func (d *AccountCreditDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{Attributes: map[string]datasourceschema.Attribute{"credit": datasourceschema.StringAttribute{Computed: true}}}
}
func (d *AccountCreditDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client, _ = req.ProviderData.(*Client)
	}
}
func (d *AccountCreditDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var out creditResponse
	if err := d.client.Do(ctx, "checkAccountCredit", map[string]any{}, &out); err != nil {
		resp.Diagnostics.AddError("Check account credit failed", err.Error())
		return
	}
	resp.State.Set(ctx, &creditModel{Credit: types.StringValue(out.Credit)})
}
