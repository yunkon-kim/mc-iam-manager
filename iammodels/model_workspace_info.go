/*
 * MCIAMManager API 명세서
 *
 * MCIAMManager API 명세서
 *
 * API version: v1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package iammodels

import "mc_iam_manager/models"

type WorkspaceInfo struct {
	WorkspaceId   string         `json:"workspaceId,omitempty"`
	WorkspaceName string         `json:"workspaceName,omitempty"`
	Description   string         `json:"description,omitempty"`
	ProjectList   ProjectInfos   "omitempty"
	UserList      []UserRoleInfo `json:"userList,omitempty"`
}

type WorkspaceInfos []WorkspaceInfo

func WorkspaceToWorkspaceInfo(workspace models.MCIamWorkspace) WorkspaceInfo {
	return WorkspaceInfo{
		WorkspaceId:   workspace.ID.String(),
		WorkspaceName: workspace.Name,
		Description:   workspace.Description,
	}
}

func WorkspaceListToWorkspaceInfoList(workspaces models.MCIamWorkspaces) WorkspaceInfos {
	var infos WorkspaceInfos
	for _, i := range workspaces {
		infos = append(infos, WorkspaceToWorkspaceInfo(i))
	}

	return infos
}