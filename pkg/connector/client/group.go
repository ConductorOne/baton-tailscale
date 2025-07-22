package client

import (
	"context"
	"errors"
	"slices"

	"github.com/conductorone/baton-tailscale/pkg/connutils"
	"github.com/tailscale/hujson"
)

func GetGroupRulesFromHujson(input hujson.ValueTrimmed, groupName string) ([]string, error) {
	rootObj, ok := input.(*hujson.Object)
	if !ok {
		return nil, errors.New("root value was not an object")
	}
	for _, member := range rootObj.Members {
		name, err := connutils.GetObjectMemberName(member)
		if err != nil {
			return nil, err
		}
		if name != "groups" {
			continue
		}

		groupList, ok := member.Value.Value.(*hujson.Object)
		if !ok {
			return nil, errors.New("groups was not an object")
		}
		for _, groupMember := range groupList.Members {
			name, err = connutils.GetObjectMemberName(groupMember)
			if err != nil {
				return nil, err
			}
			if name != groupName {
				continue
			}

			return connutils.GetPatternFromHujson(
				groupMember.Value.Value,
				connutils.IsValidEmail,
			), nil
		}

		break
	}

	return []string{}, nil
}

func FindGroupArray(input *hujson.Value, groupName string) (*hujson.Array, error) {
	rootObj, ok := input.Value.(*hujson.Object)
	if !ok {
		return nil, errors.New("root value was not an object")
	}
	for _, member := range rootObj.Members {
		name, err := connutils.GetObjectMemberName(member)
		if err != nil {
			return nil, err
		}
		if name != "groups" {
			continue
		}

		groupList, ok := member.Value.Value.(*hujson.Object)
		if !ok {
			return nil, errors.New("groups was not an object")
		}
		for _, groupMember := range groupList.Members {
			name, err = connutils.GetObjectMemberName(groupMember)
			if err != nil {
				return nil, err
			}
			if name != groupName {
				continue
			}

			groupUserList, ok := groupMember.Value.Value.(*hujson.Array)
			if !ok {
				return nil, errors.New("group list was not an array")
			}
			return groupUserList, nil
		}

		break
	}

	return nil, errors.New("group arrays not found")
}

// AddEmailToGroup TODO MARCOS FIRST DESCRIBE
func AddEmailToGroup(
	ctx context.Context,
	input *hujson.Value,
	groupName string,
	email string,
) (bool, error) {
	groupUserList, err := FindGroupArray(input, groupName)
	if err != nil {
		return false, err
	}
	defer input.Format()

	emails := connutils.Convert(
		groupUserList.Elements,
		func(in hujson.Value) string {
			lit, ok := in.Value.(hujson.Literal)
			if !ok {
				return ""
			}
			return lit.String()
		},
	)

	if slices.Contains(emails, email) {
		return false, nil
	}

	groupUserList.Elements = append(
		groupUserList.Elements,
		hujson.ArrayElement{
			Value: hujson.String(email),
		},
	)

	return true, nil
}

func RemoveEmailFromGroup(
	ctx context.Context,
	input *hujson.Value,
	groupName string,
	email string,
) (bool, error) {
	groupUserList, err := FindGroupArray(input, groupName)
	if err != nil {
		return false, err
	}

	wasRemoved := false
	for i := 0; i < len(groupUserList.Elements); i++ {
		element := groupUserList.Elements[i]
		literalEmail, ok := element.Value.(hujson.Literal)
		if !ok {
			return wasRemoved, errors.New("expected email but wasn't a literal")
		}

		if literalEmail.String() == email {
			groupUserList.Elements = append(
				groupUserList.Elements[:i],
				groupUserList.Elements[i+1:]...,
			)
			wasRemoved = true
			i--
		}
	}

	input.Format()
	return wasRemoved, nil
}
