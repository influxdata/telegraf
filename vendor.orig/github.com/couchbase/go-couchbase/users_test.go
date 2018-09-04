package couchbase

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestGetRolesAll(t *testing.T) {
	client, err := ConnectWithAuthCreds("http://localhost:8091", "Administrator", "password")
	if err != nil {
		t.Fatalf("Unable to connect: %v", err)
	}
	roles, err := client.GetRolesAll()
	if err != nil {
		t.Fatalf("Unable to get roles: %v", err)
	}

	cases := make(map[string]RoleDescription, 2)
	cases["admin"] = RoleDescription{Role: "admin", Name: "Admin", Desc: "Can manage ALL cluster features including security.", Ce: true}
	cases["query_select"] = RoleDescription{Role: "query_select", BucketName: "*", Name: "Query Select",
		Desc: "Can execute SELECT statement on bucket to retrieve data"}
	for roleName, expectedValue := range cases {
		foundThisRole := false
		for _, foundValue := range roles {
			if foundValue.Role == roleName {
				foundThisRole = true
				if expectedValue == foundValue {
					break // OK for this role
				}
				t.Fatalf("Unexpected value for role %s. Expected %+v, got %+v", roleName, expectedValue, foundValue)
			}
		}
		if !foundThisRole {
			t.Fatalf("Could not find role %s", roleName)
		}
	}
}

func TestUserUnmarshal(t *testing.T) {
	text := `[{"id":"ivanivanov","name":"Ivan Ivanov","roles":[{"role":"cluster_admin"},{"bucket_name":"default","role":"bucket_admin"}]},
			{"id":"petrpetrov","name":"Petr Petrov","roles":[{"role":"replication_admin"}]}]`
	users := make([]User, 0)

	err := json.Unmarshal([]byte(text), &users)
	if err != nil {
		t.Fatalf("Unable to unmarshal: %v", err)
	}

	expected := []User{
		User{Id: "ivanivanov", Name: "Ivan Ivanov", Roles: []Role{
			Role{Role: "cluster_admin"},
			Role{Role: "bucket_admin", BucketName: "default"}}},
		User{Id: "petrpetrov", Name: "Petr Petrov", Roles: []Role{
			Role{Role: "replication_admin"}}},
	}
	if !reflect.DeepEqual(users, expected) {
		t.Fatalf("Unexpected unmarshalled result. Expected %v, got %v.", expected, users)
	}

	ivanRoles := rolesToParamFormat(users[0].Roles)
	ivanRolesExpected := "cluster_admin,bucket_admin[default]"
	if ivanRolesExpected != ivanRoles {
		t.Errorf("Unexpected param for Ivan. Expected %v, got %v.", ivanRolesExpected, ivanRoles)
	}
	petrRoles := rolesToParamFormat(users[1].Roles)
	petrRolesExpected := "replication_admin"
	if petrRolesExpected != petrRoles {
		t.Errorf("Unexpected param for Petr. Expected %v, got %v.", petrRolesExpected, petrRoles)
	}

}
