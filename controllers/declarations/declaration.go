package declarations

import (
	"opms/controllers"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/utils/pagination"
	"os"
	"strconv"
	"strings"
	. "opms/models/declarations"
	"time"
	"opms/utils"
	. "opms/models/users"
	. "opms/models/messages"
)

type ManagerDeclarationController struct {
	controllers.BaseController
}

func (this *ManagerDeclarationController) Get() {
	//权限检测

	if !strings.Contains(this.GetSession("userPermission").(string), "declaration-manage") {
		this.Abort("401")
	}
	page, err := this.GetInt("p")
	status := this.GetString("status")
	result := this.GetString("result")

	if err != nil {
		page = 1
	}

	offset, err1 := beego.AppConfig.Int("pageoffset")
	if err1 != nil {
		offset = 15
	}

	condArr := make(map[string]string)
	condArr["status"] = status
	condArr["result"] = result
	condArr["userid"] = fmt.Sprintf("%d", this.BaseController.UserUserId)

	countDeclaration:= CountDeclaration(condArr)

	paginator := pagination.SetPaginator(this.Ctx, offset, countDeclaration)
	_, _, declarations := ListDeclaration(condArr, page, offset)

	this.Data["paginator"] = paginator
	this.Data["condArr"] = condArr
	this.Data["declarations"] = declarations
	this.Data["countdeclaration"] = countDeclaration

	this.TplName = "declarations/index.tpl"
}

type ApprovalDeclarationController struct {
	controllers.BaseController
}

func (this *ApprovalDeclarationController) Get() {
	//权限检测
	if !strings.Contains(this.GetSession("userPermission").(string), "declaration-approval") {
		this.Abort("401")
	}
	page, err := this.GetInt("p")
	filter := this.GetString("filter")

	if err != nil {
		page = 1
	}

	offset, err1 := beego.AppConfig.Int("pageoffset")
	if err1 != nil {
		offset = 15
	}

	condArr := make(map[string]string)
	condArr["filter"] = filter
	if filter == "over" {
		condArr["status"] = "1"
	} else {
		condArr["filter"] = "wait"
		condArr["status"] = "0"
	}
	condArr["userid"] = fmt.Sprintf("%d", this.BaseController.UserUserId)

	countDeclaration := CountDeclarationApproval(condArr)

	paginator := pagination.SetPaginator(this.Ctx, offset, countDeclaration)
	_, _, declarations := ListDeclarationApproval(condArr, page, offset)

	this.Data["paginator"] = paginator
	this.Data["condArr"] = condArr
	this.Data["declarations"] = declarations
	this.Data["countdeclaration"] = countDeclaration
	this.TplName = "declarations/approval.tpl"
}


type AddDeclarationController struct {
	controllers.BaseController
}

func (this *AddDeclarationController) Get() {
	//权限检测
	if !strings.Contains(this.GetSession("userPermission").(string), "declaration-add") {
		this.Abort("401")
	}
	var declaration Declarations
	this.Data["declaration"] = declaration

	_, _, users := ListUserFind()
	this.Data["users"] = users

	this.TplName = "declarations/form.tpl"
}

func (this *AddDeclarationController) Post() {
	//权限检测
	if !strings.Contains(this.GetSession("userPermission").(string), "declaration-add") {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "无权设置"}
		this.ServeJSON()
		return
	}

	objects := make([]string, 0)
	this.Ctx.Input.Bind(&objects, "objects")
	if len(objects) < 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "请填写申报项目"}
		this.ServeJSON()
		return
	}

	types := make([]string, 0)
	this.Ctx.Input.Bind(&types, "types")
	if len(types) < 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "请填写申报类型"}
		this.ServeJSON()
		return
	}
	contents := make([]string, 0)
	this.Ctx.Input.Bind(&contents, "contents")
	if len(contents) < 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "请填写申报明细"}
		this.ServeJSON()
		return
	}

	approverids := strings.Trim(this.GetString("approverid"), ",")
	if "" == approverids {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "请选择审核批人"}
		this.ServeJSON()
		return
	}

	var filepath string
	f, h, err := this.GetFile("picture")

	if err == nil {
		defer f.Close()

		//生成上传路径
		now := time.Now()
		dir := "./static/uploadfile/" + strconv.Itoa(now.Year()) + "-" + strconv.Itoa(int(now.Month())) + "/" + strconv.Itoa(now.Day())
		err1 := os.MkdirAll(dir, 0755)
		if err1 != nil {
			this.Data["json"] = map[string]interface{}{"code": 1, "message": "目录权限不够"}
			this.ServeJSON()
			return
		}
		filename := h.Filename
		if err != nil {
			this.Data["json"] = map[string]interface{}{"code": 0, "message": err}
			this.ServeJSON()
			return
		} else {
			this.SaveToFile("picture", dir+"/"+filename)
			filepath = strings.Replace(dir, ".", "", 1) + "/" + filename
		}
	}

	var declaration Declarations
	declarationid := utils.SnowFlakeId()
	declaration.Id = declarationid
	declaration.Userid = this.BaseController.UserUserId
	declaration.Objects = strings.Join(objects, "||")
	declaration.Types = strings.Join(types, "||")
	declaration.Contents = strings.Join(contents, "||")
	declaration.Files = filepath
	declaration.Approverids = approverids

	err = AddDeclaration(declaration)

	if err == nil {
		//审批人入库
		var declarationApp DeclarationsApprover
		userids := strings.Split(approverids, ",")
		for _, v := range userids {
			userid, _ := strconv.Atoi(v)
			id := utils.SnowFlakeId()
			declarationApp.Id = id
			declarationApp.Userid = int64(userid)
			declarationApp.Declarationid = declarationid
			AddDeclarationsApprover(declarationApp)
		}

		this.Data["json"] = map[string]interface{}{"code": 1, "message": "添加成功。请‘我的申报单’中设置为正常，审批人才可以看到", "id": fmt.Sprintf("%d", declarationid)}
	} else {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "申报添加失败"}
	}
	this.ServeJSON()
}

type ShowDeclarationController struct {
	controllers.BaseController
}

func (this *ShowDeclarationController) Get() {
	//权限检测
	if !strings.Contains(this.GetSession("userPermission").(string), "declaration-view") {
		this.Abort("401")
	}
	idstr := this.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idstr)
	declaration, err := GetDeclaration(int64(id))
	if err != nil {
		this.Abort("404")

	}
	this.Data["declaration"] = declaration
	_, _, approvers := ListDeclarationApproverProcess(declaration.Id)
	this.Data["approvers"] = approvers

	if this.BaseController.UserUserId != declaration.Userid {

		//检测是否可以审批和是否已审批过
		checkApproverid, checkStatus := CheckDeclarationApprover(declaration.Id, this.BaseController.UserUserId)
		if 0 == checkApproverid {
			this.Abort("401")
		}
		this.Data["checkStatus"] = checkStatus
		this.Data["checkApproverid"] = checkApproverid

		//检测审批顺序
		var checkApproverCan = 1
		for i, v := range approvers {
			if v.Status == 2 {
				checkApproverCan = 0
				break
			}
			if v.Userid == this.BaseController.UserUserId {
				if i != 0 {
					if approvers[i-1].Status == 0 {
						checkApproverCan = 0
						break
					}
				}
			}
		}
		this.Data["checkApproverCan"] = checkApproverCan

	} else {
		this.Data["checkStatus"] = 0
		this.Data["checkApproverCan"] = 0
	}

	this.TplName = "declarations/detail.tpl"
}

func (this *ShowDeclarationController) Post() {
	//权限检测
	if !strings.Contains(this.GetSession("userPermission").(string), "declaration-approval") {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "无权设置"}
		this.ServeJSON()
		return
	}

	approverid, _ := this.GetInt64("id")
	if approverid <= 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "参数出错"}
		this.ServeJSON()
		return
	}

	declarationid, _ := this.GetInt64("declarationid")
	if declarationid <= 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "参数出错"}
		this.ServeJSON()
		return
	}

	status, _ := this.GetInt("status")
	if status <= 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "请选择状态"}
		this.ServeJSON()
		return
	}
	summary := this.GetString("summary")

	var declaration DeclarationsApprover
	declaration.Status = status
	declaration.Summary = summary
	declaration.Declarationid = declarationid
	err := UpdateDeclarationsApprover(approverid, declaration)

	if err == nil {
		//消息通知
		exp, _ := GetDeclaration(declarationid)
		var msg Messages
		msg.Id = utils.SnowFlakeId()
		msg.Userid = this.BaseController.UserUserId
		msg.Touserid = exp.Userid
		msg.Type = 3
		msg.Subtype = 33
		if status == 1 {
			msg.Title = "同意"
		} else if status == 2 {
			msg.Title = "拒绝"
		}
		msg.Url = "/declaration/approval/" + fmt.Sprintf("%d", declarationid)
		AddMessages(msg)
		this.Data["json"] = map[string]interface{}{"code": 1, "message": "审批成功"}
	} else {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "审批失败"}
	}
	this.ServeJSON()
}

type EditDeclarationController struct {
	controllers.BaseController
}

func (this *EditDeclarationController) Get() {
	//权限检测
	if !strings.Contains(this.GetSession("userPermission").(string), "declaration-edit") {
		this.Abort("401")
	}
	idstr := this.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idstr)
	declaration, _ := GetDeclaration(int64(id))

	if declaration.Userid != this.BaseController.UserUserId {
		this.Abort("401")
	}
	if declaration.Status != 1 {
		this.Abort("401")
	}

	/*
		type Test struct {
			Typep   string
			Content string
			object  string
		}
		var test []Test
		objects := strings.Split(declaration.Objects, "||")
		var objectsMap = make(map[int]float64)
		for i, v := range objects {
			object, _ := strconv.ParseFloat(v, 64)
			objectsMap[i] = object
		}

		types := strings.Split(declaration.Types, "||")
		var typesMap = make(map[int]string)
		for i, v := range types {
			typesMap[i] = v
		}

		contents := strings.Split(declaration.Contents, "||")
		var contentsMap = make(map[int]string)
		for i, v := range contents {
			contentsMap[i] = v
		}
		this.Data["objectsMap"] = objectsMap
		this.Data["typesMap"] = typesMap
		this.Data["contentsMap"] = contentsMap
	*/
	this.Data["declaration"] = declaration
	this.TplName = "declarations/form.tpl"
}
func (this *EditDeclarationController) Post() {
	//权限检测
	if !strings.Contains(this.GetSession("userPermission").(string), "declaration-edit") {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "无权设置"}
		this.ServeJSON()
		return
	}
	id, _ := this.GetInt64("id")
	if id <= 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "参数出错"}
		this.ServeJSON()
		return
	}

	objects := make([]string, 0)
	this.Ctx.Input.Bind(&objects, "objects")
	if len(objects) < 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "请填写申报金额"}
		this.ServeJSON()
		return
	}
	var total float64
	for _, v := range objects {
		tmp, _ := strconv.ParseFloat(v, 64)
		total = total + tmp
	}

	types := make([]string, 0)
	this.Ctx.Input.Bind(&types, "types")
	if len(types) < 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "请填写申报类型"}
		this.ServeJSON()
		return
	}
	contents := make([]string, 0)
	this.Ctx.Input.Bind(&contents, "contents")
	if len(contents) < 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "请填写申报明细"}
		this.ServeJSON()
		return
	}
	var filepath string
	f, h, err := this.GetFile("picture")

	if err == nil {
		defer f.Close()

		//生成上传路径
		now := time.Now()
		dir := "./static/uploadfile/" + strconv.Itoa(now.Year()) + "-" + strconv.Itoa(int(now.Month())) + "/" + strconv.Itoa(now.Day())
		err1 := os.MkdirAll(dir, 0755)
		if err1 != nil {
			this.Data["json"] = map[string]interface{}{"code": 1, "message": "目录权限不够"}
			this.ServeJSON()
			return
		}
		filename := h.Filename
		if err != nil {
			this.Data["json"] = map[string]interface{}{"code": 0, "message": err}
			this.ServeJSON()
			return
		} else {
			this.SaveToFile("picture", dir+"/"+filename)
			filepath = strings.Replace(dir, ".", "", 1) + "/" + filename
		}
	}

	var declaration Declarations
	declaration.Objects = strings.Join(objects, "||")
	declaration.Types = strings.Join(types, "||")
	declaration.Contents = strings.Join(contents, "||")
	declaration.Files = filepath

	err = UpdateDeclaration(id, declaration)

	if err == nil {
		this.Data["json"] = map[string]interface{}{"code": 1, "message": "申报修改成功", "id": fmt.Sprintf("%d", id)}
	} else {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "申报修改失败"}
	}
	this.ServeJSON()
}


type AjaxDeclarationDeleteController struct {
	controllers.BaseController
}

func (this *AjaxDeclarationDeleteController) Post() {
	//权限检测
	if !strings.Contains(this.GetSession("userPermission").(string), "declaration-edit") {
		this.Abort("401")
	}
	id, _ := this.GetInt64("id")
	if id <= 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "参数出错"}
		this.ServeJSON()
		return
	}
	declaration, _ := GetDeclaration(int64(id))

	if declaration.Userid != this.BaseController.UserUserId {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "无权操作"}
		this.ServeJSON()
		return
	}
	if declaration.Status != 1 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "申报状态修改成正常，不能再删除"}
		this.ServeJSON()
		return
	}
	err := DeleteDeclaration(id)

	if err == nil {
		this.Data["json"] = map[string]interface{}{"code": 1, "message": "删除成功"}
	} else {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "删除失败"}
	}
	this.ServeJSON()
}

type AjaxDeclarationStatusController struct {
	controllers.BaseController
}

func (this *AjaxDeclarationStatusController) Post() {
	//权限检测
	if !strings.Contains(this.GetSession("userPermission").(string), "declaration-edit") {
		this.Abort("401")
	}
	id, _ := this.GetInt64("id")
	if id <= 0 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "参数出错"}
		this.ServeJSON()
		return
	}
	declaration, _ := GetDeclaration(int64(id))

	if declaration.Userid != this.BaseController.UserUserId {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "无权操作"}
		this.ServeJSON()
		return
	}
	if declaration.Status != 1 {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "申报状态修改成正常，不能再删除"}
		this.ServeJSON()
		return
	}
	err := ChangeDeclarationStatus(id, 2)

	if err == nil {
		userids := strings.Split(declaration.Approverids, ",")
		for _, v := range userids {
			//消息通知
			userid, _ := strconv.Atoi(v)
			var msg Messages
			msg.Id = utils.SnowFlakeId()
			msg.Userid = this.BaseController.UserUserId
			msg.Touserid = int64(userid)
			msg.Type = 4
			msg.Subtype = 33
			msg.Title = "去审批处理"
			msg.Url = "/declaration/approval/" + fmt.Sprintf("%d", declaration.Id)
			AddMessages(msg)
		}
		this.Data["json"] = map[string]interface{}{"code": 1, "message": "状态修改成功"}
	} else {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "状态修改失败"}
	}
	this.ServeJSON()
}