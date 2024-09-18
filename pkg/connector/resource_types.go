package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

// The user resource type is for all user objects from the database.
var (
	userResourceType = &v2.ResourceType{
		Id:          "user",
		DisplayName: "User",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
	}

	groupResourceType = &v2.ResourceType{
		Id:          "group",
		DisplayName: "Group",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
	}

	sshRuleResourceType = &v2.ResourceType{
		Id:          "sshrule",
		DisplayName: "SSH Rule",
	}

	aclRuleResourceType = &v2.ResourceType{
		Id:          "aclrule",
		DisplayName: "ACL Rule",
	}

	roleResourceType = &v2.ResourceType{
		Id:          "role",
		DisplayName: "Role",
	}
)
