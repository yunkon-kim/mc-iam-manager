package handler

import (
	"mc_iam_manager/models"
	"net/http"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

func MappingWsUserRole(tx *pop.Connection, bindModel *models.MCIamWsUserRoleMappings) (models.MCIamWsUserRoleMappings, error) {
	if bindModel != nil {
		for _, mapping := range *bindModel {
			wsUserMapper := &models.MCIamWsUserRoleMapping{}

			wsId := mapping.WsID
			roleId := mapping.RoleID
			userId := mapping.UserID

			q := tx.Eager().Where("ws_id = ?", wsId)
			q = q.Where("role_id = ?", roleId)
			q = q.Where("user_id = ?", userId)

			b, err := q.Exists(wsUserMapper)
			if err != nil {
				cblogger.Error("Error bind, ?, ?, ?", wsId, userId, roleId)
				return nil, err
			}

			if b {
				cblogger.Error("Already mapped ws user, ?, ?, ?", wsId, userId, roleId)
				return nil, errors.Wrap(err, "Already mapped ws user")
			}
		}

		LogPrintHandler("mapping ws user role bind model", bindModel)

		err := tx.Create(bindModel)

		if err != nil {
			cblogger.Error("Create Err, ?", err)
			return nil, err
		}
	}

	return *bindModel, nil

}

func MappingWsUser(tx *pop.Connection, bindModel *models.MCIamWsUserMapping) map[string]interface{} {
	if bindModel != nil {
		wsUserModel := &models.MCIamWsUserRoleMapping{}

		wsId := bindModel.WsID
		userId := bindModel.UserID

		q := tx.Eager().Where("ws_id = ?", wsId)
		q = q.Where("user_id = ?", userId)

		b, err := q.Exists(wsUserModel)
		if err != nil {
			return map[string]interface{}{
				"error":  "something query error",
				"status": "301",
			}
		}

		if b {
			return map[string]interface{}{
				"error":  "already Exists",
				"status": "301",
			}
		}
	}
	LogPrintHandler("mapping ws user bind model", bindModel)
	err := tx.Create(bindModel)

	if err != nil {
		return map[string]interface{}{
			"message": err,
			"status":  http.StatusBadRequest,
		}
	}
	return map[string]interface{}{
		"message": "success",
		"status":  http.StatusOK,
	}
}

func GetWsUserRole(tx *pop.Connection, userId string) *models.MCIamWsUserRoleMappings {

	respModel := &models.MCIamWsUserRoleMappings{}

	if userId != "" {
		q := tx.Eager().Where("user_id = ?", userId)
		err := q.All(respModel)
		if err != nil {

		}
	}

	// if role_id := bindModel.RoleID; role_id != uuid.Nil {
	// 	q := tx.Eager().Where("role_id = ?", role_id)
	// 	err := q.All(respModel)
	// 	if err != nil {

	// 	}
	// }
	// if ws_id := bindModel.WsID; ws_id != uuid.Nil {
	// 	q := tx.Eager().Where("ws_id = ?", ws_id)
	// 	err := q.All(respModel)
	// 	if err != nil {

	// 	}
	// }
	return respModel
}

func AttachProjectToWorkspace(tx *pop.Connection, bindModel models.MCIamWsProjectMappings) map[string]interface{} {

	wsPjModel := &models.MCIamWsProjectMapping{}
	for _, obj := range bindModel {
		wsPjModel.ProjectID = obj.ProjectID
		wsPjModel.WsID = obj.WsID

		q := tx.Eager().Where("ws_id = ?", obj.WsID)
		q = q.Where("project_id = ?", obj.ProjectID)
		b, err := q.Exists(wsPjModel)
		if err != nil {
			return map[string]interface{}{
				"message": "something query error",
				"status":  "301",
			}
		}

		if b {
			return map[string]interface{}{
				"message": "already Exists",
				"status":  "301",
			}
		}

		LogPrintHandler("mapping ws project bind model", wsPjModel)

		//workspace 존재 여부 체크
		wsQuery := models.DB.Where("id = ?", obj.WsID)
		existWs, err := wsQuery.Exists(models.MCIamWorkspace{})
		if !existWs {
			cblogger.Error("Workspace not exist, WSID : ", obj.WsID)
			return map[string]interface{}{
				"message": "Workspace not exist, WSID : " + obj.WsID.String(),
				"status":  http.StatusBadRequest,
			}
		}

		//project 존재 여부 체크
		projectQuery := models.DB.Where("id = ?", obj.ProjectID)
		existPj, err := projectQuery.Exists(models.MCIamProject{})
		if !existPj {
			cblogger.Error("Project not exist, PjId : ", obj.ProjectID)
			return map[string]interface{}{
				"message": "Project not exist, PjId : " + obj.ProjectID.String(),
				"status":  http.StatusBadRequest,
			}
		}

		err2 := tx.Create(wsPjModel)

		if err2 != nil {
			LogPrintHandler("mapping ws project error", err)

			return map[string]interface{}{
				"message": err,
				"Mapping": wsPjModel,
				"status":  http.StatusBadRequest,
			}
		}
	}

	return map[string]interface{}{
		"message": "success",
		"status":  http.StatusOK,
	}
}

func MappingGetProjectByWorkspace(wsId string) (models.ParserWsProjectMapping, error) {
	ws := &models.MCIamWsProjectMappings{}
	parsingWs := &models.ParserWsProjectMapping{}
	cblogger.Info("wsId : ", wsId)
	wsQuery := models.DB.Eager().Where("ws_id =?", wsId)
	projects, err := wsQuery.Exists(ws)

	cblogger.Info("projects:", projects)

	if err != nil {
		cblogger.Error(err)
		return models.ParserWsProjectMapping{}, err
	}

	if projects {
		err2 := wsQuery.All(ws)

		if err2 != nil {
			cblogger.Error(err)
			return models.ParserWsProjectMapping{}, err
		}

		parsingWs = ParserWsProjectByWs(*ws, wsId)
	}

	return *parsingWs, nil

}

func MappingWsProjectValidCheck(tx *pop.Connection, wsId string, projectId string) map[string]interface{} {
	ws := &models.MCIamWsProjectMapping{}

	q := tx.Eager().Where("ws_id =?", wsId)
	q = q.Where("project_id =?", projectId)

	b, _ := q.Exists(ws)
	// if err != nil {
	// 	return map[string]interface{}{
	// 		"error":  "something query error",
	// 		"status": "301",
	// 	}
	// }

	if b {
		project := GetProject(tx, projectId)
		return map[string]interface{}{
			"message": "valid",
			"project": project,
		}
	}
	return map[string]interface{}{
		"message": "invalid",
		"error":   "invalid project",
		"status":  "301",
	}

}

func ParserWsProjectByWs(bindModels []models.MCIamWsProjectMapping, ws_id string) *models.ParserWsProjectMapping {
	parserWsProject := &models.ParserWsProjectMapping{}
	projectArray := models.MCIamProjects{}
	wsUuid, _ := uuid.FromString(ws_id)
	cblogger.Info("#### bindmodels ####", bindModels)
	for _, obj := range bindModels {
		cblogger.Info("#### wsuuid ####", obj.WsID)
		if wsUuid == obj.WsID {
			parserWsProject.WsID = obj.WsID
			parserWsProject.Ws = obj.Ws
			if obj.ProjectID != uuid.Nil {
				projectArray = append(projectArray, *obj.Project)
				parserWsProject.Projects = projectArray
			}
		}

	}

	cblogger.Info("parserWsProject : ", parserWsProject)
	return parserWsProject
}

// func ParserWsProject(tx *pop.Connection, bindModels []models.MCIamWorkspace) *models.ParserWsProjectMappings {
// 	parserWsProject := []models.ParserWsProjectMapping{}
// 	projectArray := []models.MCIamProject{}

// 	ParserWsProjectByWs

// 	return parserWsProject
// }

func MappingUserRole(tx *pop.Connection, bindModel *models.MCIamUserRoleMapping) map[string]interface{} {
	if bindModel.ID != uuid.Nil {
		userRoleModel := &models.MCIamUserRoleMapping{}

		roleId := bindModel.RoleID
		userId := bindModel.UserID

		q := tx.Eager().Where("role_id = ?", roleId)
		q = q.Where("user_id = ?", userId)

		b, err := q.Exists(userRoleModel)
		if err != nil {
			return map[string]interface{}{
				"error":  "something query error",
				"status": "301",
			}
		}

		if b {
			return map[string]interface{}{
				"error":  "already Exists",
				"status": "301",
			}
		}
	}
	LogPrintHandler("mapping user role bind model", bindModel)
	err := tx.Create(bindModel)

	if err != nil {
		return map[string]interface{}{
			"message": err,
			"status":  http.StatusBadRequest,
		}
	}
	return map[string]interface{}{
		"message": "success",
		"status":  http.StatusOK,
	}
}

func MappingDeleteWsProject(tx *pop.Connection, bindModel *models.MCIamWsProjectMapping) map[string]interface{} {
	err := tx.Destroy(bindModel)
	if err != nil {
		return map[string]interface{}{
			"message": errors.WithStack(err),
			"status":  "301",
		}
	}
	return map[string]interface{}{
		"message": "success",
		"status":  http.StatusOK,
	}

}
