package github

type ConfigOpts func(c *CfgParams) error

type CfgParams struct {
	Token string
}

func OptsToken(token string) ConfigOpts {
	return func(c *CfgParams) error {
		c.Token = token
		return nil
	}
}

// ----- getters

func (c *CfgParams) GetToken() string {
	return c.Token
}
