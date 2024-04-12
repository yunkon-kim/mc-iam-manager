/*
 * MCIAMManager API 명세서
 *
 * MCIAMManager API 명세서
 *
 * API version: v1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package iammodels

import (
	"mc_iam_manager/models"
)

type ProjectInfo struct {
	ProjectId   string `json:"projectId,omitempty"`
	ProjectName string `json:"projectName,omitempty"`
	Description string `json:"description,omitempty"`
}
type ProjectInfos []ProjectInfo

func ProjectToProjectInfo(project models.MCIamProject) ProjectInfo {
	return ProjectInfo{
		project.ID.String(),
		project.Name,
		project.Description,
	}
}

func ProjectsToProjectInfoList(projects models.MCIamProjects) ProjectInfos {
	var infos []ProjectInfo
	for _, i := range projects {
		infos = append(infos, ProjectToProjectInfo(i))
	}

	return infos
}
