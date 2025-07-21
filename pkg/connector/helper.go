package connector

import (
	"strconv"

	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-tailscale/pkg/connector/client"
)

func unmarshalSkipToken(token *pagination.Token) (int32, *pagination.Bag, error) {
	b := &pagination.Bag{}
	err := b.Unmarshal(token.Token)
	if err != nil {
		return 0, nil, err
	}
	current := b.Current()
	skip := int32(0)
	if current != nil && current.Token != "" {
		skip64, err := strconv.ParseInt(current.Token, 10, 32)
		if err != nil {
			return 0, nil, err
		}
		skip = int32(skip64)
	}
	return skip, b, nil
}

func GetUserIDsFromUserEmails(users []client.User, emails []string) []string {
	IDperEmail := make(map[string]string)
	for _, user := range users {
		IDperEmail[user.LoginName] = user.ID
	}

	var userIDs []string
	for _, email := range emails {
		if _, found := IDperEmail[email]; found {
			userIDs = append(userIDs, IDperEmail[email])
		}
	}

	return userIDs
}
