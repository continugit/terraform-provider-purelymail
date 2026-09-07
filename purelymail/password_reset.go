package purelymail

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resetModel struct {
	Username      types.String `tfsdk:"username"`
	Type          types.String `tfsdk:"type"`
	Target        types.String `tfsdk:"target"`
	Description   types.String `tfsdk:"description"`
	AllowMFAReset types.Bool   `tfsdk:"allow_mfa_reset"`
}

func resetAttrs() map[string]resourceschema.Attribute {
	return map[string]resourceschema.Attribute{
		"username":        resourceschema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"type":            resourceschema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"target":          resourceschema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"description":     resourceschema.StringAttribute{Optional: true},
		"allow_mfa_reset": resourceschema.BoolAttribute{Optional: true},
	}
}

type PasswordResetResource struct{ client *Client }

func NewPasswordResetResource() resource.Resource { return &PasswordResetResource{} }
func (r *PasswordResetResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "purelymail_password_reset_method"
}
func (r *PasswordResetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{Attributes: resetAttrs()}
}
func (r *PasswordResetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client, _ = req.ProviderData.(*Client)
	}
}
func (r *PasswordResetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var p resetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsert(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, &p)
}
func (r *PasswordResetResource) upsert(ctx context.Context, p resetModel, diags *diag.Diagnostics) {
	body := map[string]any{"userName": p.Username.ValueString(), "type": p.Type.ValueString(), "target": p.Target.ValueString(), "description": p.Description.ValueString(), "allowMfaReset": p.AllowMFAReset.ValueBool()}
	if err := r.client.Do(ctx, "upsertPasswordReset", body, &Empty{}); err != nil {
		diags.AddError("Upsert password reset method failed", err.Error())
	}
}
func (r *PasswordResetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var s resetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &s)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out listResetResponse
	if err := r.client.Do(ctx, "listPasswordReset", map[string]string{"userName": s.Username.ValueString()}, &out); err != nil {
		resp.Diagnostics.AddError("Read password reset method failed", err.Error())
		return
	}
	for _, item := range out.Users {
		if item.Type == s.Type.ValueString() && item.Target == s.Target.ValueString() {
			s.Description = types.StringValue(item.Description)
			s.AllowMFAReset = types.BoolValue(item.AllowMFAReset)
			resp.State.Set(ctx, &s)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *PasswordResetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var p resetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsert(ctx, p, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.State.Set(ctx, &p)
	}
}
func (r *PasswordResetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var s resetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &s)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Do(ctx, "deletePasswordReset", map[string]string{"userName": s.Username.ValueString(), "target": s.Target.ValueString()}, &Empty{}); err != nil {
		resp.Diagnostics.AddError("Delete password reset method failed", err.Error())
	}
}

type PasswordResetsDataSource struct{ client *Client }
type resetListModel struct {
	Username types.String `tfsdk:"username"`
	Methods  []resetModel `tfsdk:"methods"`
}

func NewPasswordResetsDataSource() datasource.DataSource { return &PasswordResetsDataSource{} }
func (d *PasswordResetsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "purelymail_password_reset_methods"
}
func (d *PasswordResetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{Attributes: map[string]datasourceschema.Attribute{"username": datasourceschema.StringAttribute{Required: true}, "methods": datasourceschema.ListNestedAttribute{Computed: true, NestedObject: datasourceschema.NestedAttributeObject{Attributes: map[string]datasourceschema.Attribute{"type": datasourceschema.StringAttribute{Computed: true}, "target": datasourceschema.StringAttribute{Computed: true}, "description": datasourceschema.StringAttribute{Computed: true}, "allow_mfa_reset": datasourceschema.BoolAttribute{Computed: true}}}}}}
}
func (d *PasswordResetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client, _ = req.ProviderData.(*Client)
	}
}
func (d *PasswordResetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m resetListModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out listResetResponse
	if err := d.client.Do(ctx, "listPasswordReset", map[string]string{"userName": m.Username.ValueString()}, &out); err != nil {
		resp.Diagnostics.AddError("List password reset methods failed", err.Error())
		return
	}
	m.Methods = make([]resetModel, len(out.Users))
	for i, item := range out.Users {
		m.Methods[i] = resetModel{Type: types.StringValue(item.Type), Target: types.StringValue(item.Target), Description: types.StringValue(item.Description), AllowMFAReset: types.BoolValue(item.AllowMFAReset)}
	}
	resp.State.Set(ctx, &m)
}

func resetID(username, typ, target string) string {
	return strings.Join([]string{username, typ, target}, "|")
}
