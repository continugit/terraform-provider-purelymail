package purelymail

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type routingModel struct {
	ID              types.Int64    `tfsdk:"id"`
	DomainName      types.String   `tfsdk:"domain_name"`
	Prefix          types.Bool     `tfsdk:"prefix"`
	MatchUser       types.String   `tfsdk:"match_user"`
	TargetAddresses []types.String `tfsdk:"target_addresses"`
	Catchall        types.Bool     `tfsdk:"catchall"`
}

func routingAttrs() map[string]resourceschema.Attribute {
	return map[string]resourceschema.Attribute{
		"id":               resourceschema.Int64Attribute{Computed: true},
		"domain_name":      resourceschema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"prefix":           resourceschema.BoolAttribute{Required: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()}},
		"match_user":       resourceschema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"target_addresses": resourceschema.ListAttribute{Required: true, ElementType: types.StringType},
		"catchall":         resourceschema.BoolAttribute{Optional: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()}},
	}
}

func setRouting(m *routingModel, r RoutingRule) {
	m.ID = types.Int64Value(r.ID)
	m.DomainName = types.StringValue(r.DomainName)
	m.Prefix = types.BoolValue(r.Prefix)
	m.MatchUser = types.StringValue(r.MatchUser)
	m.Catchall = types.BoolValue(r.Catchall)
	m.TargetAddresses = make([]types.String, len(r.TargetAddresses))
	for i, a := range r.TargetAddresses {
		m.TargetAddresses[i] = types.StringValue(a)
	}
}

type RoutingRuleResource struct{ client *Client }

func NewRoutingRuleResource() resource.Resource { return &RoutingRuleResource{} }
func (r *RoutingRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "purelymail_routing_rule"
}
func (r *RoutingRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{Attributes: routingAttrs()}
}
func (r *RoutingRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client, _ = req.ProviderData.(*Client)
	}
}
func (r *RoutingRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *RoutingRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var p routingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}
	targets := make([]string, len(p.TargetAddresses))
	for i, a := range p.TargetAddresses {
		targets[i] = a.ValueString()
	}
	body := map[string]any{"domainName": p.DomainName.ValueString(), "prefix": p.Prefix.ValueBool(), "matchUser": p.MatchUser.ValueString(), "targetAddresses": targets, "catchall": p.Catchall.ValueBool()}
	if err := r.client.Do(ctx, "createRoutingRule", body, &Empty{}); err != nil {
		resp.Diagnostics.AddError("Create routing rule failed", err.Error())
		return
	}
	if err := r.find(ctx, &p); err != nil {
		resp.Diagnostics.AddError("Read routing rule after create failed", err.Error())
		return
	}
	resp.State.Set(ctx, &p)
}
func (r *RoutingRuleResource) find(ctx context.Context, m *routingModel) error {
	var out listRulesResponse
	if err := r.client.Do(ctx, "listRoutingRules", map[string]any{}, &out); err != nil {
		return err
	}
	for _, rule := range out.Rules {
		if rule.DomainName == m.DomainName.ValueString() && rule.Prefix == m.Prefix.ValueBool() && rule.MatchUser == m.MatchUser.ValueString() {
			setRouting(m, rule)
			return nil
		}
	}
	return fmt.Errorf("rule was not returned by the API")
}
func (r *RoutingRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var s routingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &s)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.findByID(ctx, &s); err != nil {
		resp.Diagnostics.AddError("Read routing rule failed", err.Error())
		return
	}
	resp.State.Set(ctx, &s)
}
func (r *RoutingRuleResource) findByID(ctx context.Context, m *routingModel) error {
	var out listRulesResponse
	if err := r.client.Do(ctx, "listRoutingRules", map[string]any{}, &out); err != nil {
		return err
	}
	for _, rule := range out.Rules {
		if rule.ID == m.ID.ValueInt64() {
			setRouting(m, rule)
			return nil
		}
	}
	return fmt.Errorf("routing rule %d no longer exists", m.ID.ValueInt64())
}
func (r *RoutingRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var p routingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var s routingModel
	req.State.Get(ctx, &s)
	if err := r.client.Do(ctx, "deleteRoutingRule", map[string]int64{"routingRuleId": s.ID.ValueInt64()}, &Empty{}); err != nil {
		resp.Diagnostics.AddError("Replace routing rule failed", err.Error())
		return
	}
	targets := make([]string, len(p.TargetAddresses))
	for i, a := range p.TargetAddresses {
		targets[i] = a.ValueString()
	}
	body := map[string]any{"domainName": p.DomainName.ValueString(), "prefix": p.Prefix.ValueBool(), "matchUser": p.MatchUser.ValueString(), "targetAddresses": targets, "catchall": p.Catchall.ValueBool()}
	if err := r.client.Do(ctx, "createRoutingRule", body, &Empty{}); err != nil {
		resp.Diagnostics.AddError("Recreate routing rule failed", err.Error())
		return
	}
	if err := r.find(ctx, &p); err != nil {
		resp.Diagnostics.AddError("Read recreated routing rule failed", err.Error())
		return
	}
	resp.State.Set(ctx, &p)
}
func (r *RoutingRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var s routingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &s)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Do(ctx, "deleteRoutingRule", map[string]int64{"routingRuleId": s.ID.ValueInt64()}, &Empty{}); err != nil {
		resp.Diagnostics.AddError("Delete routing rule failed", err.Error())
	}
}

type RoutingRulesDataSource struct{ client *Client }
type routingListModel struct {
	Rules []routingModel `tfsdk:"rules"`
}

func NewRoutingRulesDataSource() datasource.DataSource { return &RoutingRulesDataSource{} }
func (d *RoutingRulesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "purelymail_routing_rules"
}
func (d *RoutingRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{Attributes: map[string]datasourceschema.Attribute{"rules": datasourceschema.ListNestedAttribute{Computed: true, NestedObject: datasourceschema.NestedAttributeObject{Attributes: map[string]datasourceschema.Attribute{"id": datasourceschema.Int64Attribute{Computed: true}, "domain_name": datasourceschema.StringAttribute{Computed: true}, "prefix": datasourceschema.BoolAttribute{Computed: true}, "match_user": datasourceschema.StringAttribute{Computed: true}, "target_addresses": datasourceschema.ListAttribute{Computed: true, ElementType: types.StringType}, "catchall": datasourceschema.BoolAttribute{Computed: true}}}}}}
}
func (d *RoutingRulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client, _ = req.ProviderData.(*Client)
	}
}
func (d *RoutingRulesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var out listRulesResponse
	if err := d.client.Do(ctx, "listRoutingRules", map[string]any{}, &out); err != nil {
		resp.Diagnostics.AddError("List routing rules failed", err.Error())
		return
	}
	m := routingListModel{Rules: make([]routingModel, len(out.Rules))}
	for i, rule := range out.Rules {
		setRouting(&m.Rules[i], rule)
	}
	resp.State.Set(ctx, &m)
}
