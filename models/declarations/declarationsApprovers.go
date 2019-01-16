package declarations

import (
//"fmt"
"opms/models"
//"opms/utils"
"time"

//"github.com/astaxie/beego"
"github.com/astaxie/beego/orm"
)

type DeclarationsApprover struct {
	Id        int64 `orm:"pk;column(approverid);"`
	Declarationid int64
	Userid    int64
	Summary   string
	Status    int
	Created   int64
	Changed   int64
}

func (this *DeclarationsApprover) TableName() string {
	return models.TableName("declarations_approver")
}

func init() {
	orm.RegisterModel(new(DeclarationsApprover))
}

func AddDeclarationsApprover(upd DeclarationsApprover) error {
	o := orm.NewOrm()
	declaration := new(DeclarationsApprover)

	declaration.Id = upd.Id
	declaration.Userid = upd.Userid
	declaration.Declarationid = upd.Declarationid
	//declaration.Summary = upd.Summary
	declaration.Status = 0
	declaration.Created = time.Now().Unix()
	declaration.Changed = time.Now().Unix()
	_, err := o.Insert(declaration)
	return err
}

func UpdateDeclarationsApprover(id int64, upd DeclarationsApprover) error {
	var declaration DeclarationsApprover
	o := orm.NewOrm()
	declaration = DeclarationsApprover{Id: id}

	declaration.Summary = upd.Summary
	declaration.Status = upd.Status
	declaration.Changed = time.Now().Unix()
	_, err := o.Update(&declaration, "summary", "status", "changed")
	if err == nil {
		//直接结束
		if upd.Status == 2 {
			ChangeDeclarationResult(upd.Declarationid, 2)
			o.Raw("UPDATE pms_declarations_approver SET status = ?,summary = ?, changed = ? WHERE declarationid = ? AND approverid != ?", 2, "前审批人拒绝，后面审批人默认为拒绝状态", time.Now().Unix(), upd.Declarationid, id).Exec()
		} else {
			_, _, approvers := ListDeclarationApproverProcess(upd.Declarationid)
			//检测审批顺序
			var ApproverNum = 0
			for _, v := range approvers {
				if v.Status == 1 {
					ApproverNum++
				}
			}
			if ApproverNum == len(approvers) {
				ChangeDeclarationResult(upd.Declarationid, 1)
			}
		}
	}
	return err
}

type DeclarationApproverProcess struct {
	Userid   int64
	Realname string
	Avatar   string
	Position string
	Status   int
	Summary  string
	Changed  int64
}

func ListDeclarationApproverProcess(declarationid int64) (num int64, err error, user []DeclarationApproverProcess) {
	var users []DeclarationApproverProcess
	qb, _ := orm.NewQueryBuilder("mysql")
	qb.Select("upr.userid", "upr.realname", "p.name AS position", "u.avatar", "la.status", "la.summary", "la.changed").From("pms_declarations_approver AS la").
		LeftJoin("pms_users AS u").On("u.userid = la.userid").
		LeftJoin("pms_users_profile AS upr").On("upr.userid = u.userid").
		LeftJoin("pms_positions AS p").On("p.positionid = upr.positionid").
		Where("la.declarationid=?").
		OrderBy("la.approverid").
		Asc()
	sql := qb.String()
	o := orm.NewOrm()
	nums, err := o.Raw(sql, declarationid).QueryRows(&users)
	return nums, err, users
}

func ListDeclarationApproverProcessHtml(declarationid int64) string {
	nums, _, users := ListDeclarationApproverProcess(declarationid)
	var html, avatar, css, status string
	var num = int(nums)
	for i, v := range users {
		if "" == v.Avatar {
			avatar = "/static/img/avatar/1.jpg"
		} else {
			avatar = v.Avatar
		}
		if v.Status == 1 {
			status = "同意"
		} else if v.Status == 2 {
			//css = "gray"
			status = "拒绝"
		} else {
			css = "gray"
			status = "未处"
		}

		html += "<a href='javascript:;' title='" + v.Realname + "'><img class='" + css + "' src='" + avatar + "' alt='" + v.Realname + "'>" + status + "</a>"
		if i < (num - 1) {
			html += "<span>..........</span>"
		}
	}
	return html
}

//检测是否已经审批
func CheckDeclarationApprover(id, userId int64) (int64, int) {
	var declaration DeclarationsApprover
	o := orm.NewOrm()
	o.QueryTable(models.TableName("declarations_approver")).Filter("declarationid", id).Filter("userid", userId).One(&declaration, "approverid", "status")

	return declaration.Id, declaration.Status
}
