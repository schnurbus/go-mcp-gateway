package auth

func (a *Auth) GetAuthorizationPath() string {
	return a.authorizePath
}

func (a *Auth) GetAuthorizationURL() string {
	return a.baseURL + a.authorizePath
}

func (a *Auth) GetCallbackPath() string {
	return a.callbackPath
}

func (a *Auth) GetCallbackURL() string {
	return a.baseURL + a.callbackPath
}

func (a *Auth) GetDynamicRegistrationPath() string {
	return a.registerPath
}

func (a *Auth) GetDynamicRegistrationURL() string {
	return a.baseURL + a.registerPath
}

func (a *Auth) GetSupportCodeChallengeMethods() []string {
	return a.supportedCodeChallengeMethods
}

func (a *Auth) GetSupportGrantTypes() []string {
	return a.supportedGrantTypes
}

func (a *Auth) GetSupportResponseTypes() []string {
	return a.supportedResponseTypes
}

func (a *Auth) GetSupportTokenEndpointAuthMethods() []string {
	return a.supportedTokenEndpointAuthMethods
}

func (a *Auth) GetTokenPath() string {
	return a.tokenPath
}

func (a *Auth) GetTokenURL() string {
	return a.baseURL + a.tokenPath
}
