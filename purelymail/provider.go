package purelymail

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Provider struct {
	version string
}

type providerData struct {
	APIToken types.String `tfsdk:"api_token"`
	BaseURL  types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &Provider{version: version} }
}

func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "purelymail"
	response.Version = p.version
}

func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = providerschema.Schema{
		Attributes: map[string]providerschema.Attribute{
			"api_token": providerschema.StringAttribute{
				Optional: true, Sensitive: true,
				Description: "Purelymail API token. May also be set with PURELYMAIL_API_TOKEN.",
			},
			"base_url": providerschema.StringAttribute{
				Optional: true, Description: "Purelymail API base URL.",
			},
		},
	}
}

func (p *Provider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var data providerData
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	token := data.APIToken.ValueString()
	if data.APIToken.IsNull() || data.APIToken.ValueString() == "" {
		token = os.Getenv("PURELYMAIL_API_TOKEN")
	}
	if token == "" {
		response.Diagnostics.AddError("Missing Purelymail API token", "Set api_token or PURELYMAIL_API_TOKEN.")
		return
	}
	baseURL := data.BaseURL.ValueString()
	if data.BaseURL.IsNull() || baseURL == "" {
		baseURL = "https://purelymail.com"
	}
	response.DataSourceData = NewClient(baseURL, token)
	response.ResourceData = NewClient(baseURL, token)
}

func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainResource, NewUserResource, NewRoutingRuleResource, NewPasswordResetResource,
	}
}

func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainsDataSource, NewDomainDataSource, NewOwnershipCodeDataSource,
		NewUsersDataSource, NewUserDataSource, NewRoutingRulesDataSource,
		NewPasswordResetsDataSource, NewAccountCreditDataSource,
	}
}

var _ provider.Provider = (*Provider)(nil)
