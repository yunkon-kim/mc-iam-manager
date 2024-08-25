package keycloak

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Nerzal/gocloak/v13"
	"github.com/m-cmp/mc-iam-manager/handler"
)

var (
	kc Keycloak
)

func init() {
	kc = Keycloak{
		KcClient:     gocloak.NewClient(handler.KEYCLOAK_HOST),
		Host:         handler.KEYCLOAK_HOST,
		Realm:        handler.KEYCLAOK_REALM,
		Client:       handler.KEYCLAOK_CLIENT,
		ClientSecret: handler.KEYCLAOK_CLIENT_SECRET,
	}
}

func KeycloakGetCerts() (*gocloak.CertResponse, error) {
	ctx := context.Background()
	cert, err := kc.KcClient.GetCerts(ctx, kc.Realm)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return cert, nil
}

// realm-management manage-clients role 필요
func KeycloakGetClientInfo(accessToken string) (*gocloak.Client, error) {
	ctx := context.Background()
	clinetResp, err := kc.KcClient.GetClientRepresentation(ctx, accessToken, kc.Realm, kc.Client)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return clinetResp, nil
}

// Auth Management

func KeycloakLogin(id string, password string) (*gocloak.JWT, error) {
	ctx := context.Background()
	accessTokenResponse, err := kc.KcClient.Login(ctx, kc.Client, kc.ClientSecret, kc.Realm, id, password)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return accessTokenResponse, nil
}

func KeycloakRefreshToken(refreshToken string) (*gocloak.JWT, error) {
	ctx := context.Background()
	accessTokenResponse, err := kc.KcClient.RefreshToken(ctx, refreshToken, kc.Client, kc.ClientSecret, kc.Realm)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return accessTokenResponse, nil
}

func KeycloakLogout(refreshToken string) error {
	ctx := context.Background()
	err := kc.KcClient.Logout(ctx, kc.Client, kc.ClientSecret, kc.Realm, refreshToken)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func KeycloakGetUserInfo(accessToken string) (*gocloak.UserInfo, error) {
	ctx := context.Background()
	userinfo, err := kc.KcClient.GetUserInfo(ctx, accessToken, kc.Realm)
	if err != nil {
		log.Println(err)
	}
	return userinfo, nil
}

func KeycloakTokenInfo(accessToken string) (*gocloak.IntroSpectTokenResult, error) {
	ctx := context.Background()
	userinfo, err := kc.KcClient.RetrospectToken(ctx, accessToken, kc.Client, kc.ClientSecret, kc.Realm)
	if err != nil {
		log.Println(err)
	}
	return userinfo, nil
}

// Users Management
type CreateUserRequset struct {
	Name      string `json:"id"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

func KeycloakCreateUser(accessToken string, user CreateUserRequset, password string) error {
	ctx := context.Background()

	userInfo := gocloak.User{
		Username:  &user.Name,
		Enabled:   gocloak.BoolP(true),
		FirstName: &user.FirstName,
		LastName:  &user.LastName,
		Email:     &user.Email,
	}

	userUUID, err := kc.KcClient.CreateUser(ctx, accessToken, kc.Realm, userInfo)
	if err != nil {
		log.Println(err)
		return err
	}

	err = kc.KcClient.SetPassword(ctx, accessToken, userUUID, kc.Realm, password, false)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func KeycloakGetUsers(accessToken string, userId string) ([]*gocloak.User, error) {
	ctx := context.Background()

	userInfo := gocloak.GetUsersParams{
		Username: &userId,
	}

	users, err := kc.KcClient.GetUsers(ctx, accessToken, kc.Realm, userInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return users, nil
}

func KeycloakDeleteUser(accessToken string, userId string) error {
	ctx := context.Background()

	err := kc.KcClient.DeleteUser(ctx, accessToken, kc.Realm, userId)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func KeycloakUpdateUser(accessToken string, user CreateUserRequset, userUUID string) error {
	ctx := context.Background()

	userInfo := gocloak.User{
		ID:        &userUUID,
		FirstName: &user.FirstName,
		LastName:  &user.LastName,
		Email:     &user.Email,
	}

	err := kc.KcClient.UpdateUser(ctx, accessToken, kc.Realm, userInfo)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Resource Management

func KeycloakCreateResources(accessToken string, resources CreateResourceRequestArr) (*[]gocloak.ResourceRepresentation, error) {
	ctx := context.Background()

	//realm-management manage-clients role 필요
	clinetResp, err := KeycloakGetClientInfo(accessToken)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result := []gocloak.ResourceRepresentation{}
	createResourceerrors := []error{}
	for _, resource := range resources {
		name := resource.Framework + ":res:" + resource.OperationId + ":" + resource.Method + ":" + resource.URI
		URIS := []string{resource.URI}
		resreq := gocloak.ResourceRepresentation{
			Name: &name,
			URIs: &URIS,
		}
		res, err := kc.KcClient.CreateResource(ctx, accessToken, kc.Realm, *clinetResp.ID, resreq)
		if err != nil {
			log.Println(err)
			createResourceerrors = append(createResourceerrors, err)
			continue
		}
		result = append(result, *res)
	}

	if len(createResourceerrors) != 0 {
		return nil, errors.Join(createResourceerrors...)
	}

	return &result, nil
}

func KeycloakCreateMenuResources(accessToken string, resources CreateMenuResourceRequestArr) (*[]gocloak.ResourceRepresentation, error) {
	ctx := context.Background()

	//realm-management manage-clients role 필요
	clinetResp, err := KeycloakGetClientInfo(accessToken)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	result := []gocloak.ResourceRepresentation{}
	createResourceerrors := []error{}
	for _, resource := range resources {
		resreq := gocloak.ResourceRepresentation{
			Name: gocloak.StringP(resource.Framework + ":menu:" + resource.Id + ":" + resource.DisplayName + ":" + resource.ParentMenuId + ":" + resource.Priority + ":" + resource.IsAction),
		}
		res, err := kc.KcClient.CreateResource(ctx, accessToken, kc.Realm, *clinetResp.ID, resreq)
		if err != nil {
			log.Println(err)
			createResourceerrors = append(createResourceerrors, err)
			continue
		}
		result = append(result, *res)
	}

	if len(createResourceerrors) != 0 {
		return nil, errors.Join(createResourceerrors...)
	}

	return &result, nil
}

func KeycloakGetResources(accessToken string, params gocloak.GetResourceParams) ([]*gocloak.ResourceRepresentation, error) {
	ctx := context.Background()
	resource, err := kc.KcClient.GetResourcesClient(ctx, accessToken, kc.Realm, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return resource, nil
}

func KeycloakUpdateResources(accessToken string, resourceid string, resource CreateResourceRequest) error {
	ctx := context.Background()

	//realm-management manage-clients role 필요
	clientResp, err := KeycloakGetClientInfo(accessToken)
	if err != nil {
		log.Println(err)
		return err
	}
	name := resource.Framework + ":" + resource.OperationId + ":" + resource.Method + ":" + resource.URI
	URIS := []string{resource.URI}
	resreq := gocloak.ResourceRepresentation{
		ID:   &resourceid,
		Name: &name,
		URIs: &URIS,
	}
	err = kc.KcClient.UpdateResource(ctx, accessToken, kc.Realm, *clientResp.ID, resreq)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func KeycloakDeleteResource(accessToken string, resourceid string) error {
	ctx := context.Background()

	//realm-management manage-clients role 필요
	clientResp, err := KeycloakGetClientInfo(accessToken)
	if err != nil {
		log.Println(err)
		return err
	}

	err = kc.KcClient.DeleteResource(ctx, accessToken, kc.Realm, *clientResp.ID, resourceid)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Role Management

// realm-management manage-realm role require
func KeycloakCreateRole(accessToken string, name string, desc string) (string, error) {
	ctx := context.Background()

	rolereq := gocloak.Role{
		Name:        &name,
		Description: &desc,
	}

	res, err := kc.KcClient.CreateRealmRole(ctx, accessToken, kc.Realm, rolereq)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return res, nil
}

func KeycloakGetRole(accessToken string, name string) (*gocloak.Role, error) {
	ctx := context.Background()

	res, err := kc.KcClient.GetRealmRole(ctx, accessToken, kc.Realm, name)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return res, nil
}

func keycloakGetUserRoles(accessToken string, userUUID string) ([]*gocloak.Role, error) {
	ctx := context.Background()

	roles, err := kc.KcClient.GetRealmRolesByUserID(ctx, accessToken, kc.Realm, userUUID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return roles, nil
}

func KeycloakUpdateRole(accessToken string, name string, desc string) error {
	ctx := context.Background()

	rolereq := gocloak.Role{
		Name:        &name,
		Description: &desc,
	}

	err := kc.KcClient.UpdateRealmRole(ctx, accessToken, kc.Realm, name, rolereq)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func KeycloakDeleteRole(accessToken string, name string) error {
	ctx := context.Background()

	err := kc.KcClient.DeleteRealmRole(ctx, accessToken, kc.Realm, name)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func KeycloakMappingUserRole(accessToken string, userId string, roleReq string) error {
	ctx := context.Background()

	userInfo := gocloak.GetUsersParams{
		Exact:    gocloak.BoolP(true),
		Username: &userId,
	}
	user, err := kc.KcClient.GetUsers(ctx, accessToken, kc.Realm, userInfo)
	if err != nil {
		log.Println(err)
		return err
	}
	if len(user) != 0 && *user[0].Username != userId {
		err := fmt.Errorf("%s user not found ", userId)
		log.Println(err)
		return err
	}

	role, err := KeycloakGetRole(accessToken, roleReq)
	if err != nil {
		log.Println(err)
		return err
	}

	inputRoles := []gocloak.Role{*role}
	err = kc.KcClient.AddRealmRoleToUser(ctx, accessToken, kc.Realm, *user[0].ID, inputRoles)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func KeycloakUnMappingUserRole(accessToken string, userId string, roleReq string) error {
	ctx := context.Background()

	userInfo := gocloak.GetUsersParams{
		Exact:    gocloak.BoolP(true),
		Username: &userId,
	}
	user, err := kc.KcClient.GetUsers(ctx, accessToken, kc.Realm, userInfo)
	if err != nil {
		log.Println(err)
		return err
	}
	if len(user) != 0 && *user[0].Username != userId {
		err := fmt.Errorf("%s user not found ", userId)
		log.Println(err)
		return err
	}

	role, err := KeycloakGetRole(accessToken, roleReq)
	if err != nil {
		log.Println(err)
		return err
	}

	inputRoles := []gocloak.Role{*role}
	err = kc.KcClient.DeleteRealmRoleFromUser(ctx, accessToken, kc.Realm, *user[0].ID, inputRoles)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Policy Management

func KeycloakCreatePolicy(accessToken string, name string, desc string) (*gocloak.PolicyRepresentation, error) {
	ctx := context.Background()

	//realm-management manage-clients role 필요
	clinetResp, err := KeycloakGetClientInfo(accessToken)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	role, err := KeycloakGetRole(accessToken, name)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	roles := []gocloak.RoleDefinition{{
		ID: role.ID,
	}}

	policyName := name + "Policy"
	policyDesc := desc + " Policy"
	policyType := "role"

	policyreq := gocloak.PolicyRepresentation{
		Name:        &policyName,
		Description: &policyDesc,
		Type:        &policyType,
		RolePolicyRepresentation: gocloak.RolePolicyRepresentation{
			Roles: &roles,
		},
	}

	res, err := kc.KcClient.CreatePolicy(ctx, accessToken, kc.Realm, *clinetResp.ID, policyreq)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return res, nil
}

func KeycloakDeletePolicy(accessToken string, policyId string) error {
	ctx := context.Background()

	clinetResp, err := KeycloakGetClientInfo(accessToken)
	if err != nil {
		log.Println(err)
		return err
	}

	err = kc.KcClient.DeletePolicy(ctx, accessToken, kc.Realm, *clinetResp.ID, policyId)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Permission Management

type permissionDetail struct {
	Permission *gocloak.PermissionRepresentation `json:"permission"`
	Resources  []*gocloak.PermissionResource     `json:"resources"`
	Policies   []*gocloak.PolicyRepresentation   `json:"rolePolicies"`
}

func KeycloakCreatePermission(accessToken string, name string, desc string, permissionResources []string, permissionPolicies []string) (*gocloak.PermissionRepresentation, error) {
	ctx := context.Background()

	//realm-management manage-clients role 필요
	clinetResp, err := KeycloakGetClientInfo(accessToken)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	permissionName := name + "Permission"
	permissionDesc := desc + " Permission"
	permissionType := "resource"
	permissionReq := gocloak.PermissionRepresentation{
		Name:             &permissionName,
		Description:      &permissionDesc,
		Type:             &permissionType,
		Resources:        &permissionResources,
		Policies:         &permissionPolicies,
		DecisionStrategy: gocloak.AFFIRMATIVE,
	}

	res, err := kc.KcClient.CreatePermission(ctx, accessToken, kc.Realm, *clinetResp.ID, permissionReq)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return res, nil
}

func KeycloakGetPermissions(accessToken string, reqParam gocloak.GetPermissionParams) ([]*gocloak.PermissionRepresentation, error) {
	ctx := context.Background()

	//realm-management manage-clients role 필요
	clinetResp, err := KeycloakGetClientInfo(accessToken)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := kc.KcClient.GetPermissions(ctx, accessToken, kc.Realm, *clinetResp.ID, reqParam)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return res, nil
}

func KeycloakGetPermissionDetail(accessToken string, id string) (*permissionDetail, error) {
	ctx := context.Background()

	//realm-management manage-clients role 필요
	clinetResp, err := KeycloakGetClientInfo(accessToken)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	permissionRes, err := kc.KcClient.GetPermission(ctx, accessToken, kc.Realm, *clinetResp.ID, id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	resourcesRes, err := kc.KcClient.GetPermissionResources(ctx, accessToken, kc.Realm, *clinetResp.ID, id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	policyRes, err := kc.KcClient.GetAuthorizationPolicyAssociatedPolicies(ctx, accessToken, kc.Realm, *clinetResp.ID, *permissionRes.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result := &permissionDetail{}
	result.Permission = permissionRes
	result.Resources = resourcesRes
	result.Policies = policyRes

	return result, nil
}

func KeycloakUpdatePermission(accessToken string, id string, name string, desc string, permissionResources []string, permissionPolicies []string) error {
	ctx := context.Background()

	//realm-management manage-clients role 필요
	clinetResp, err := KeycloakGetClientInfo(accessToken)
	if err != nil {
		log.Println(err)
		return err
	}

	permissionType := "resource"
	permissionReq := gocloak.PermissionRepresentation{
		ID:               &id,
		Name:             &name,
		Description:      &desc,
		Type:             &permissionType,
		Resources:        &permissionResources,
		Policies:         &permissionPolicies,
		DecisionStrategy: gocloak.AFFIRMATIVE,
	}

	err = kc.KcClient.UpdatePermission(ctx, accessToken, kc.Realm, *clinetResp.ID, permissionReq)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func KeycloakDeletePermission(accessToken string, id string) error {
	ctx := context.Background()

	//realm-management manage-clients role 필요
	clinetResp, err := KeycloakGetClientInfo(accessToken)
	if err != nil {
		log.Println(err)
		return err
	}

	err = kc.KcClient.DeletePermission(ctx, accessToken, kc.Realm, *clinetResp.ID, id)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Ticket Management

func KeycloakGetTicketByRequestUri(accessToken string, uri []string) (*gocloak.JWT, error) {
	ctx := context.Background()

	opt := gocloak.RequestingPartyTokenOptions{
		GrantType:                     gocloak.StringP("urn:ietf:params:oauth:grant-type:uma-ticket"),
		Audience:                      gocloak.StringP(kc.Client),
		Permissions:                   &uri,
		PermissionResourceFormat:      gocloak.StringP("uri"),
		PermissionResourceMatchingURI: gocloak.BoolP(true),
	}

	ticket, err := kc.KcClient.GetRequestingPartyToken(ctx, accessToken, kc.Realm, opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return ticket, nil
}

func KeycloakGetAvaliablePermission(accessToken string) (*[]gocloak.RequestingPartyPermission, error) {
	ctx := context.Background()

	opt := gocloak.RequestingPartyTokenOptions{
		GrantType:    gocloak.StringP("urn:ietf:params:oauth:grant-type:uma-ticket"),
		Audience:     gocloak.StringP(kc.Client),
		ResponseMode: gocloak.StringP("permissions"),
	}

	ticket, err := kc.KcClient.GetRequestingPartyPermissions(ctx, accessToken, kc.Realm, opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return ticket, nil
}

func KeycloakGetTicketByFrameworkResourceName(accessToken string, framework string, name string) (*gocloak.JWT, error) {
	ctx := context.Background()

	params := gocloak.GetResourceParams{
		Name: gocloak.StringP(framework + ":res:" + name),
	}
	resources, err := KeycloakGetResources(accessToken, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(resources) == 0 {
		return nil, fmt.Errorf("resource Not Found")
	}

	nameArr := []string{*resources[0].Name}
	opt := gocloak.RequestingPartyTokenOptions{
		GrantType:   gocloak.StringP("urn:ietf:params:oauth:grant-type:uma-ticket"),
		Audience:    gocloak.StringP(kc.Client),
		Permissions: &nameArr,
		// PermissionResourceFormat: gocloak.StringP("id"),
		// PermissionResourceMatchingURI: gocloak.BoolP(true),
	}
	ticket, err := kc.KcClient.GetRequestingPartyToken(ctx, accessToken, kc.Realm, opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return ticket, nil
}

func KeycloakGetAvailableMenus(accessToken string) (*[]gocloak.RequestingPartyPermission, error) {
	ctx := context.Background()

	params := gocloak.GetResourceParams{
		Name: gocloak.StringP("web:menu:"),
	}
	resources, err := KeycloakGetResources(accessToken, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(resources) == 0 {
		return nil, fmt.Errorf("menu Not Found")
	}

	names := make([]string, len(resources))
	for i, resource := range resources {
		names[i] = *resource.Name
	}

	opt := gocloak.RequestingPartyTokenOptions{
		GrantType:    gocloak.StringP("urn:ietf:params:oauth:grant-type:uma-ticket"),
		Audience:     gocloak.StringP(kc.Client),
		ResponseMode: gocloak.StringP("permissions"),
		Permissions:  &names,
	}
	ticket, err := kc.KcClient.GetRequestingPartyPermissions(ctx, accessToken, kc.Realm, opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return ticket, nil
}