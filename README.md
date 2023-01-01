# Azure Demo: List all tenants a user has access to

It seems the ability to get the list of tenants are not available using the Azure SDK for Go. So, I figured it would be
a good opportunity to write a small demo on how to do it.

It's important to impersonate a user, so that way, the user can
register a new tenant, and use the app to "onboard" the new tenant. This app
doesn't do onboarding, but instead, it just lists all the users in all the
tenants the user has access to.

Onboarding tenants typically involve:

- registering an application for security monitoring (e.g. SIEM)
- registering an application for test and development (e.g. Azure DevOps or GitHub)
- registering an emergency account (e.g. for a security incident)
- registering an application for identity management (e.g. Okta)
