package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"

    "github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
    "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
    auth "github.com/microsoft/kiota-authentication-azure-go"
    msgraphsdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
)

type Tenant struct {
    Id             string   `json:"id"`
    TenantId       string   `json:"tenantId"`
    CountryCode    string   `json:"countryCode"`
    DisplayName    string   `json:"displayName"`
    Domains        []string `json:"domains"`
    TenantCategory string   `json:"tenantCategory"`
    DefaultDomain  string   `json:"defaultDomain"`
    TenantType     string   `json:"tenantType"`
}

type TenantResponse struct {
    Value []*Tenant `json:"value"`
}

var tenantID = os.Getenv("AZURE_TENANT_ID")
var clientID = os.Getenv("AZURE_CLIENT_ID")
var redirectUrl = os.Getenv("AZURE_REDIRECT_URL")

func main() {

    scopes := []string{
        // "openid",
        "https://management.azure.com/user_impersonation",
    }

    tenants, err := getTenants(scopes)
    if err != nil {
        log.Fatalln(err)
    }
    repeat := "------------------------------------"
    fmt.Printf("%-36s %-36s %-36s\n", repeat, repeat, repeat)
    fmt.Printf("%-36s %-36s %-36s\n", "Id", "TenantId", "DisplayName")
    fmt.Printf("%-36s %-36s %-36s\n", repeat, repeat, repeat)
    for _, tenant := range tenants {
        fmt.Printf("%-36s %-36s %-36s\n", tenant.DisplayName, tenant.TenantId, tenant.DefaultDomain)
    }

    for _, tenant := range tenants {

        repeat := strings.Repeat("=", 36*3+2)
        fmt.Println()
        fmt.Printf("%s\n", repeat)
        fmt.Println(tenant.DisplayName, "("+tenant.DefaultDomain+")")
        fmt.Printf("%s\n", repeat)

        err := onboard(tenant)
        if err != nil {
            log.Println("ERROR:", err.Error())
        }
    }
}

func onboard(tenant *Tenant) error {
    options := &azidentity.InteractiveBrowserCredentialOptions{
        TenantID:    tenant.TenantId,
        ClientID:    clientID,
        RedirectURL: redirectUrl,
    }

    credential, err := azidentity.NewInteractiveBrowserCredential(options)
    if err != nil {
        fmt.Println(err.Error())
        return err
    }

    var ctx = context.Background()

    scopes := []string{
        "https://graph.microsoft.com/Directory.ReadWrite.All",
    }
    // Create an auth provider using the credential
    authProvider, err := auth.NewAzureIdentityAuthenticationProviderWithScopes(credential, scopes)
    if err != nil {
        return err
    }

    // Create a request adapter using the auth provider
    adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
    if err != nil {
        return err
    }

    graphClient := msgraphsdk.NewGraphServiceClient(adapter)

    result, err := graphClient.Users().Get(ctx, nil)
    if err != nil {
        return err
    }

    repeat := strings.Repeat("-", 36*3+2)
    fmt.Printf("%s\n", repeat)
    fmt.Printf("%-36s %-36s %-36s\n", "Display Name", "Id", "Email")
    fmt.Printf("%s\n", repeat)
    users := result.GetValue()
    if len(users) > 0 {
        for _, user := range users {
            fmt.Printf("%-36s %-36s %-36s\n", s(user.GetDisplayName(), ""), s(user.GetId(), ""), s(user.GetMail(), ""))
        }
    } else {
        fmt.Println("No users found")
    }

    return nil
}

func s(value *string, defaultValue string) string {
    if value != nil {
        return *value
    }
    return defaultValue
}

func getTenants(scopes []string) (tenants []*Tenant, err error) {
    options := &azidentity.InteractiveBrowserCredentialOptions{
        TenantID:    tenantID,
        ClientID:    clientID,
        RedirectURL: redirectUrl,
    }

    credential, err := azidentity.NewInteractiveBrowserCredential(options)
    if err != nil {
        fmt.Println(err.Error())
        return nil, err
    }

    var ctx = context.Background()

    policy := policy.TokenRequestOptions{
        Scopes: scopes,
    }

    token, err := credential.GetToken(ctx, policy)
    if err != nil {

        return nil, err
    }

    bearer := "Bearer " + token.Token
    req, err := http.NewRequest(
        "GET",
        "https://management.azure.com/tenants?api-version=2020-01-01",
        bytes.NewBuffer(nil),
    )
    if err != nil {

        return nil, err
    }

    req.Header.Set("Authorization", bearer)
    req.Header.Add("Accept", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println(err.Error())
        return nil, err
    }
    defer resp.Body.Close()

    // check if the response is successful
    if resp.StatusCode != http.StatusOK {
        fmt.Println("Error: ", resp.StatusCode)

        b := new(bytes.Buffer)
        b.ReadFrom(resp.Body)
        s := b.String()
        fmt.Println(s)

        return nil, err
    }

    // deserialize the resp.Body into a TenantResponse
    decoder := json.NewDecoder(resp.Body)
    var tenantResponse TenantResponse
    err = decoder.Decode(&tenantResponse)
    if err != nil {
        fmt.Println(err.Error())
        return nil, err
    }

    return tenantResponse.Value, err
}
