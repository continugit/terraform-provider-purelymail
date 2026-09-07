package purelymail

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type userModel struct {
	Username                       types.String `tfsdk:"username"`
	Password                       types.String `tfsdk:"password"`
	EnableSearchIndexing           types.Bool   `tfsdk:"enable_search_indexing"`
	EnablePasswordReset            types.Bool   `tfsdk:"enable_password_reset"`
	RequireTwoFactorAuthentication types.Bool   `tfsdk:"require_two_factor_authentication"`
	RecoveryEnabled                types.Bool   `tfsdk:"recovery_enabled"`
	EnableSpamFiltering            types.Bool   `tfsdk:"enable_spam_filtering"`
}

type UserResource struct{ client *Client }

func NewUserResource() resource.Resource { return &UserResource{} }
func (r *UserResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "purelymail_user"
}
func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{Attributes: map[string]resourceschema.Attribute{
		"username":                          resourceschema.StringAttribute{Required: true},
		"password":                          resourceschema.StringAttribute{Optional: true, Sensitive: true},
		"enable_search_indexing":            resourceschema.BoolAttribute{Optional: true, Computed: true},
		"enable_password_reset":             resourceschema.BoolAttribute{Optional: true, Computed: true},
		"require_two_factor_authentication": resourceschema.BoolAttribute{Optional: true, Computed: true},
		"recovery_enabled":                  resourceschema.BoolAttribute{Computed: true},
		"enable_spam_filtering":             resourceschema.BoolAttribute{Computed: true},
	}}
}
func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client, _ = req.ProviderData.(*Client)
	}
}
func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("username"), req, resp)
}
func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var p userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}
	local, domain, ok := strings.Cut(p.Username.ValueString(), "@")
	if !ok {
		resp.Diagnostics.AddError("Invalid username", "username must be a full email address.")
		return
	}
	if p.Password.IsNull() || p.Password.ValueString() == "" {
		resp.Diagnostics.AddError("Missing password", "password is required when creating a user.")
		return
	}
	body := map[string]any{"userName": local, "domainName": domain, "password": p.Password.ValueString()}
	if !p.EnablePasswordReset.IsNull() {
		body["enablePasswordReset"] = p.EnablePasswordReset.ValueBool()
	}
	if err := r.client.Do(ctx, "createUser", body, &Empty{}); err != nil {
		resp.Diagnostics.AddError("Create user failed", err.Error())
		return
	}
	resp.State.Set(ctx, &p)
	r.read(ctx, p.Username.ValueString(), &p, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.State.Set(ctx, &p)
	}
}
func (r *UserResource) read(ctx context.Context, username string, state *userModel, diags *diag.Diagnostics) {
	var out UserInfo
	if err := r.client.Do(ctx, "getUser", map[string]string{"userName": username}, &out); err != nil {
		diags.AddError("Read user failed", err.Error())
		return
	}
	state.EnableSearchIndexing = types.BoolValue(out.EnableSearchIndexing)
	state.EnablePasswordReset = types.BoolValue(out.RecoveryEnabled)
	state.RecoveryEnabled = types.BoolValue(out.RecoveryEnabled)
	state.RequireTwoFactorAuthentication = types.BoolValue(out.RequireTwoFactorAuthentication)
	state.EnableSpamFiltering = types.BoolValue(out.EnableSpamFiltering)
}
func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var s userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &s)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.read(ctx, s.Username.ValueString(), &s, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.State.Set(ctx, &s)
	}
}
func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var p userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var old userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"userName": old.Username.ValueString()}
	if p.Username.ValueString() != old.Username.ValueString() {
		body["newUserName"] = p.Username.ValueString()
	}
	if !p.Password.IsNull() && !p.Password.IsUnknown() && p.Password.ValueString() != "" {
		body["newPassword"] = p.Password.ValueString()
	}
	if !p.EnableSearchIndexing.IsNull() {
		body["enableSearchIndexing"] = p.EnableSearchIndexing.ValueBool()
	}
	if !p.EnablePasswordReset.IsNull() {
		body["enablePasswordReset"] = p.EnablePasswordReset.ValueBool()
	}
	if !p.RequireTwoFactorAuthentication.IsNull() {
		body["requireTwoFactorAuthentication"] = p.RequireTwoFactorAuthentication.ValueBool()
	}
	if err := r.client.Do(ctx, "modifyUser", body, &Empty{}); err != nil {
		resp.Diagnostics.AddError("Update user failed", err.Error())
		return
	}
	r.read(ctx, p.Username.ValueString(), &p, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.State.Set(ctx, &p)
	}
}
func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var s userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &s)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Do(ctx, "deleteUser", map[string]string{"userName": s.Username.ValueString()}, &Empty{}); err != nil {
		resp.Diagnostics.AddError("Delete user failed", err.Error())
	}
}

type UsersDataSource struct{ client *Client }
type usersModel struct {
	Users []types.String `tfsdk:"users"`
}

func NewUsersDataSource() datasource.DataSource { return &UsersDataSource{} }
func (d *UsersDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "purelymail_users"
}
func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{Attributes: map[string]datasourceschema.Attribute{"users": datasourceschema.ListAttribute{Computed: true, ElementType: types.StringType}}}
}
func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client, _ = req.ProviderData.(*Client)
	}
}
func (d *UsersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var out listUsersResponse
	if err := d.client.Do(ctx, "listUser", map[string]any{}, &out); err != nil {
		resp.Diagnostics.AddError("List users failed", err.Error())
		return
	}
	m := usersModel{Users: make([]types.String, len(out.Users))}
	for i, u := range out.Users {
		m.Users[i] = types.StringValue(u)
	}
	resp.State.Set(ctx, &m)
}

type UserDataSource struct{ client *Client }
type userLookupModel struct {
	Username                       types.String `tfsdk:"username"`
	EnableSearchIndexing           types.Bool   `tfsdk:"enable_search_indexing"`
	RecoveryEnabled                types.Bool   `tfsdk:"recovery_enabled"`
	RequireTwoFactorAuthentication types.Bool   `tfsdk:"require_two_factor_authentication"`
	EnableSpamFiltering            types.Bool   `tfsdk:"enable_spam_filtering"`
}

func NewUserDataSource() datasource.DataSource { return &UserDataSource{} }
func (d *UserDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "purelymail_user"
}
func (d *UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{Attributes: map[string]datasourceschema.Attribute{"username": datasourceschema.StringAttribute{Required: true}, "enable_search_indexing": datasourceschema.BoolAttribute{Computed: true}, "recovery_enabled": datasourceschema.BoolAttribute{Computed: true}, "require_two_factor_authentication": datasourceschema.BoolAttribute{Computed: true}, "enable_spam_filtering": datasourceschema.BoolAttribute{Computed: true}}}
}
func (d *UserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client, _ = req.ProviderData.(*Client)
	}
}
func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m userLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out UserInfo
	if err := d.client.Do(ctx, "getUser", map[string]string{"userName": m.Username.ValueString()}, &out); err != nil {
		resp.Diagnostics.AddError("Get user failed", err.Error())
		return
	}
	m.EnableSearchIndexing = types.BoolValue(out.EnableSearchIndexing)
	m.RecoveryEnabled = types.BoolValue(out.RecoveryEnabled)
	m.RequireTwoFactorAuthentication = types.BoolValue(out.RequireTwoFactorAuthentication)
	m.EnableSpamFiltering = types.BoolValue(out.EnableSpamFiltering)
	resp.State.Set(ctx, &m)
}
