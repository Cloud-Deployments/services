package client

import "fmt"

func (c *Client) Login(runnerId, runnerSecret, organizationId string) error {
	apiUrl := c.buildUrl("runners", "login")

	req, err := c.newRequest(runnerId, runnerSecret, organizationId, "POST", apiUrl, nil)

	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	statusCode, _, err := req.Do()
	if err != nil {
		return fmt.Errorf("failed to login [%d]: %w", statusCode, err)
	}

	return nil
}

func (c *Client) Logout(runnerId, runnerSecret, organizationId string) error {
	apiUrl := c.buildUrl("runners", "logout")

	req, err := c.newRequest(runnerId, runnerSecret, organizationId, "POST", apiUrl, nil)

	if err != nil {
		return fmt.Errorf("failed to create logout request: %w", err)
	}

	statusCode, _, err := req.Do()
	if err != nil {
		return fmt.Errorf("failed to logout [%d]: %w", statusCode, err)
	}

	return nil
}
