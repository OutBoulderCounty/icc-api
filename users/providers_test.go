package users_test

import (
	"api/env"
	"api/users"
	"testing"
)

func TestProviderApproval(t *testing.T) {
	e := env.TestSetup(t, true, "../.env")
	selectProvider := "select id from users where approvedProvider = false"
	var providerID int64
	err := e.DB.QueryRow(selectProvider).Scan(&providerID)
	if err != nil {
		t.Error("error getting provider ID. " + err.Error())
		return
	}
	err = users.ApproveProvider(providerID, true, e.DB)
	if err != nil {
		t.Error("error approving provider. " + err.Error())
		return
	}
	// validate that the provider is now approved
	user, err := users.Get(providerID, e.DB)
	if err != nil {
		t.Error("error getting user. " + err.Error())
		return
	}
	if !user.ApprovedProvider {
		t.Error("user is not approved")
	}

	// remove approval from provider
	err = users.ApproveProvider(providerID, false, e.DB)
	if err != nil {
		t.Error("error removing approval from provider. " + err.Error())
		return
	}
}
