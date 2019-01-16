package declarations

import (
	"fmt"
	"opms/models"
	"opms/utils"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type Declarations struct {
	Id          int64 `orm:"pk;column(declarationid);"`
	Userid      int64
	Objects     string
	Types       string
	Contents    string
	Files    	string
	Result      int
	Status      int
	Approverids string
	Created     int64
	Changed     int64
}

func (this *Declarations) TableName() string {
	return models.TableName("declarations")
}

func init() {
	orm.RegisterModel(new(Declarations))
}

func AddDeclaration(upd Declarations) error {
	o := orm.NewOrm()
	declaration := new(Declarations)

	declaration.Id = upd.Id
	declaration.Userid = upd.Userid
	declaration.Objects = upd.Objects
	declaration.Types = upd.Types
	declaration.Contents = upd.Contents
	declaration.Files = upd.Files
	declaration.Status = 1
	declaration.Approverids = upd.Approverids
	declaration.Created = time.Now().Unix()
	declaration.Changed = time.Now().Unix()
	_, err := o.Insert(declaration)
	return err
}

func UpdateDeclaration(id int64, upd Declarations) error {
	var declaration Declarations
	o := orm.NewOrm()
	declaration = Declarations{Id: id}

	declaration.Objects = upd.Objects
	declaration.Types = upd.Types
	declaration.Contents = upd.Contents
	declaration.Changed = time.Now().Unix()

	var err error
	if "" != upd.Files {
		declaration.Files = upd.Files
		_, err = o.Update(&declaration, "amounts", "types", "contents", "picture", "total", "changed")
	} else {
		_, err = o.Update(&declaration, "amounts", "types", "contents", "total", "changed")
	}

	return err
}

func ListDeclaration(condArr map[string]string, page int, offset int) (num int64, err error, ops []Declarations) {
	o := orm.NewOrm()
	o.Using("default")
	qs := o.QueryTable(models.TableName("declarations"))
	cond := orm.NewCondition()

	if condArr["status"] != "" {
		cond = cond.And("status", condArr["status"])
	}
	if condArr["userid"] != "" {
		cond = cond.And("userid", condArr["userid"])
	}
	if condArr["result"] != "" {
		cond = cond.And("result", condArr["result"])
	}
	qs = qs.SetCond(cond)
	if page < 1 {
		page = 1
	}
	if offset < 1 {
		offset, _ = beego.AppConfig.Int("pageoffset")
	}
	start := (page - 1) * offset
	qs = qs.OrderBy("-declarationid")
	var declaration []Declarations
	num, errs := qs.Limit(offset, start).All(&declaration)
	return num, errs, declaration
}

func CountDeclaration(condArr map[string]string) int64 {
	o := orm.NewOrm()
	qs := o.QueryTable(models.TableName("declarations"))
	cond := orm.NewCondition()

	if condArr["status"] != "" {
		cond = cond.And("status", condArr["status"])
	}
	if condArr["userid"] != "" {
		cond = cond.And("userid", condArr["userid"])
	}
	if condArr["result"] != "" {
		cond = cond.And("result", condArr["result"])
	}
	num, _ := qs.SetCond(cond).Count()

	return num
}

//待审批
func ListDeclarationApproval(condArr map[string]string, page int, offset int) (num int64, err error, ops []Declarations) {
	if page < 1 {
		page = 1
	}
	if offset < 1 {
		offset, _ = beego.AppConfig.Int("pageoffset")
	}
	start := (page - 1) * offset
	var declaration []Declarations
	qb, _ := orm.NewQueryBuilder("mysql")
	qb.Select("l.declarationid","l.objects", "l.userid", "l.total", "l.changed", "l.approverids", "la.status").From("pms_declarations_approver AS la").
		LeftJoin("pms_declarations AS l").On("l.declarationid = la.declarationid").
		Where("la.userid=?").
		And("l.status=2")

	if condArr["status"] == "0" {
		qb.And("la.status=0")
		qb.And("l.result=0")
	} else if condArr["status"] == "1" {
		qb.And("la.status>0")
	}
	qb.OrderBy("la.approverid").
		Desc().
		Limit(offset).
		Offset(start)

	sql := qb.String()
	o := orm.NewOrm()

	nums, err := o.Raw(sql, condArr["userid"]).QueryRows(&declaration)
	return nums, err, declaration
}

type TmpDeclarationCount struct {
	Num int64
}

func CountDeclarationApproval(condArr map[string]string) int64 {
	qb, _ := orm.NewQueryBuilder("mysql")
	qb.Select("Count(1) AS num").From("pms_declarations_approver AS la").
		LeftJoin("pms_declarations AS l").On("l.declarationid = la.declarationid").
		Where("la.userid=?").
		And("l.status=2")
	if condArr["status"] == "0" {
		qb.And("la.status=0")
		qb.And("l.result=0")
	} else if condArr["status"] == "1" {
		qb.And("la.status>0")
	}
	sql := qb.String()
	o := orm.NewOrm()
	var tmp TmpDeclarationCount
	err := o.Raw(sql, condArr["userid"]).QueryRow(&tmp)
	if err == nil {
		return tmp.Num
	} else {
		return 0
	}
}

func GetDeclaration(id int64) (Declarations, error) {
	var declaration Declarations
	var err error

	err = utils.GetCache("GetDeclaration.id."+fmt.Sprintf("%d", id), &declaration)
	if err != nil {
		cache_expire, _ := beego.AppConfig.Int("cache_expire")
		o := orm.NewOrm()
		declaration = Declarations{Id: id}
		err = o.Read(&declaration)
		utils.SetCache("GetDeclaration.id."+fmt.Sprintf("%d", id), declaration, cache_expire)
	}
	return declaration, err
}

func ChangeDeclarationStatus(id int64, status int) error {
	o := orm.NewOrm()

	declaration := Declarations{Id: id}
	err := o.Read(&declaration, "declarationid")
	if nil != err {
		return err
	} else {
		declaration.Status = status
		_, err := o.Update(&declaration)
		return err
	}
}

func ChangeDeclarationResult(id int64, result int) error {
	o := orm.NewOrm()

	declaration := Declarations{Id: id}
	err := o.Read(&declaration, "declarationid")
	if nil != err {
		return err
	} else {
		declaration.Result = result
		_, err := o.Update(&declaration)
		return err
	}
}

func DeleteDeclaration(id int64) error {
	o := orm.NewOrm()
	_, err := o.Delete(&Declarations{Id: id})

	if err == nil {
		_, err = o.Raw("DELETE FROM "+models.TableName("declarations_approver")+" WHERE declarationid = ?", id).Exec()
	}
	return err
}
