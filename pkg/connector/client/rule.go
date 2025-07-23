package client

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/conductorone/baton-tailscale/pkg/connutils"
	"github.com/tailscale/hujson"
)

type ruleKey string
type rule struct {
	obj *hujson.Object
}

const (
	RuleKeySSH  ruleKey = "ssh"
	RuleKeyACLs ruleKey = "acls"
)

func (r rule) GetValueOfNamedMember(name string) []string {
	for _, member := range r.obj.Members {
		literalName, err := connutils.GetObjectMemberName(member)
		if err != nil || literalName != name {
			continue
		}

		switch v := member.Value.Value.(type) {
		case hujson.Literal:
			return []string{v.String()}
		case *hujson.Array:
			arrayMembers := []string{}
			for _, item := range v.Elements {
				arrLiteral, ok := item.Value.(hujson.Literal)
				if !ok {
					continue
				}
				arrayMembers = append(arrayMembers, arrLiteral.String())
			}
			return arrayMembers
		}
	}
	return []string{}
}

func (r rule) GetHash() string {
	action := r.GetValueOfNamedMember("action")
	dst := r.GetValueOfNamedMember("dst")
	users := r.GetValueOfNamedMember("users")

	sort.Strings(action)
	sort.Strings(dst)
	sort.Strings(users)

	actionStr := strings.Join(action, "")
	dstStr := strings.Join(dst, "")
	usersStr := strings.Join(users, "")

	return fmt.Sprintf("%x", sha256.Sum256([]byte(actionStr+dstStr+usersStr)))
}

func GetRulesFromHujson(input hujson.ValueTrimmed, ruleKey ruleKey) ([]rule, error) {
	rv := []rule{}
	rootObj, ok := input.(*hujson.Object)
	if !ok {
		return nil, errors.New("root value was not an object")
	}
	for _, member := range rootObj.Members {
		name, err := connutils.GetObjectMemberName(member)
		if err != nil {
			return nil, err
		}
		if name != string(ruleKey) {
			continue
		}

		rules, ok := member.Value.Value.(*hujson.Array)
		if !ok {
			return nil, errors.New("rule value was not an array")
		}

		for _, ruleEl := range rules.Elements {
			ruleObj, ok := ruleEl.Value.(*hujson.Object)
			if !ok {
				return nil, errors.New("rule was not an object")
			}
			rv = append(rv, rule{obj: ruleObj})
		}

		break
	}

	return rv, nil
}

// FindRuleArray TODO MARCOS DESCRIBE
func FindRuleArray(
	input *hujson.Value,
	ruleKey ruleKey,
	ruleHash string,
) (*hujson.Array, error) {
	rootObj, ok := input.Value.(*hujson.Object)
	if !ok {
		return nil, errors.New("root value was not an object")
	}

	for _, member := range rootObj.Members {
		name, err := connutils.GetObjectMemberName(member)
		if err != nil {
			return nil, err
		}
		// oneof: ssh or acls
		if name != string(ruleKey) {
			continue
		}

		rules, ok := member.Value.Value.(*hujson.Array)
		if !ok {
			return nil, errors.New("rule value was not an array")
		}

		for _, ruleElement := range rules.Elements {
			ruleObj, ok := ruleElement.Value.(*hujson.Object)
			if !ok {
				return nil, errors.New("rule was not an object")
			}

			rule := rule{obj: ruleObj}
			hash := rule.GetHash()
			if hash != ruleHash {
				continue
			}

			for _, ruleMember := range ruleObj.Members {
				ruleName, err := connutils.GetObjectMemberName(ruleMember)
				if err != nil {
					return nil, err
				}
				// If it's a 'src' array, or it's an ACL users array, remove our email
				if ruleName == "src" || (ruleName == "users" && ruleKey == RuleKeyACLs) {
					ruleMemberList, ok := ruleMember.Value.Value.(*hujson.Array)
					if !ok {
						return nil, errors.New("rule list was not an array")
					}

					return ruleMemberList, nil
				}
			}

			break
		}

		break
	}

	return nil, errors.New("no rule found for that hash")
}

func AddEmailToRule(
	ctx context.Context,
	input *hujson.Value,
	ruleKey ruleKey,
	ruleHash string,
	email string,
) (bool, error) {
	ruleArray, err := FindRuleArray(input, ruleKey, ruleHash)
	if err != nil {
		return false, err
	}
	defer input.Format()

	emails := connutils.Convert(
		ruleArray.Elements,
		func(in hujson.ArrayElement) string {
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

	ruleArray.Elements = append(
		ruleArray.Elements,
		hujson.ArrayElement{
			Value: hujson.String(email),
		},
	)

	return true, nil
}

func RemoveEmailFromRule(
	ctx context.Context,
	input *hujson.Value,
	ruleKey ruleKey,
	ruleHash string,
	email string,
) (bool, error) {
	ruleArray, err := FindRuleArray(input, ruleKey, ruleHash)
	if err != nil {
		return false, err
	}

	wasRemoved := false
	for i := 0; i < len(ruleArray.Elements); i++ {
		element := ruleArray.Elements[i]
		literalEmail, ok := element.Value.(hujson.Literal)
		if !ok {
			return wasRemoved, errors.New("expected email but wasn't a literal")
		}

		if literalEmail.String() == email {
			ruleArray.Elements = append(
				ruleArray.Elements[:i],
				ruleArray.Elements[i+1:]...,
			)
			wasRemoved = true
			i--
		}
	}

	input.Format()
	return wasRemoved, nil
}
