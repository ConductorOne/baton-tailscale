package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-tailscale/pkg/connector/client"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

func (u *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return u.resourceType
}

func userResource(ctx context.Context, user *client.User, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	var userStatus v2.UserTrait_Status_Status = v2.UserTrait_Status_STATUS_ENABLED
	firstName, lastName := rs.SplitFullName(user.DisplayName)
	profile := map[string]interface{}{
		"login":       user.LoginName,
		"first_name":  firstName,
		"last_name":   lastName,
		"email":       user.LoginName,
		"user_id":     user.ID,
		"user_status": user.Status,
	}

	switch user.Status {
	case "active":
		userStatus = v2.UserTrait_Status_STATUS_ENABLED
	case "suspended":
		userStatus = v2.UserTrait_Status_STATUS_DISABLED
	}

	userTraits := []rs.UserTraitOption{
		rs.WithUserProfile(profile),
		rs.WithStatus(userStatus),
		rs.WithUserLogin(user.LoginName),
		rs.WithEmail(user.LoginName, true),
	}

	displayName := user.DisplayName
	if displayName == "" {
		displayName = user.LoginName
	}

	ret, err := rs.NewUserResource(
		displayName,
		userResourceType,
		user.ID,
		userTraits,
		rs.WithParentResourceID(parentResourceID))
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// List always returns an empty slice, we don't sync users.
func (u *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var (
		rv            []*v2.Resource
		nextPageToken string
	)
	_, bag, err := unmarshalSkipToken(pToken)
	if err != nil {
		return nil, "", nil, err
	}

	if bag.Current() == nil {
		bag.Push(pagination.PageState{
			ResourceTypeID: userResourceType.Id,
		})
		bag.Push(pagination.PageState{
			ResourceTypeID: inviteResourceType.Id,
		})
	}

	switch bag.ResourceTypeID() {
	case userResourceType.Id:
		users, _, err := u.client.GetUsers(ctx)
		if err != nil {
			return nil, "", nil, err
		}

		err = bag.Next(nextPageToken)
		if err != nil {
			return nil, "", nil, err
		}

		for _, user := range users {
			usrCopy := user
			ur, err := userResource(ctx, &usrCopy, parentResourceID)
			if err != nil {
				return nil, "", nil, err
			}

			rv = append(rv, ur)
		}
	case inviteResourceType.Id:
		userInvites, _, err := u.client.GetUserInvites(ctx)
		if err != nil {
			return nil, "", nil, err
		}

		err = bag.Next(nextPageToken)
		if err != nil {
			return nil, "", nil, err
		}

		for _, user := range userInvites {
			ur, err := userResource(ctx, &client.User{
				ID:          user.ID,
				DisplayName: user.Email,
				LoginName:   user.Email,
				Role:        user.Role,
				Status:      "invited",
			}, parentResourceID)
			if err != nil {
				return nil, "", nil, err
			}

			rv = append(rv, ur)
		}

	default:
		return nil, "", nil, fmt.Errorf("tailscale-connector: invalid resource type: %s", bag.ResourceTypeID())
	}

	nextPageToken, err = bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return rv, nextPageToken, nil, nil
}

// Entitlements always returns an empty slice for users.
func (u *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (u *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *client.Client) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       client,
	}
}
