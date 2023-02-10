package table

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/youminxue/v2/cmd/internal/ddl/columnenum"
	"github.com/youminxue/v2/cmd/internal/ddl/extraenum"
	"github.com/youminxue/v2/cmd/internal/ddl/keyenum"
	"github.com/youminxue/v2/cmd/internal/ddl/nullenum"
	"github.com/youminxue/v2/cmd/internal/ddl/sortenum"
	"github.com/youminxue/v2/toolkit/astutils"
	"github.com/youminxue/v2/toolkit/stringutils"
	"github.com/youminxue/v2/toolkit/templateutils"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	now = "CURRENT_TIMESTAMP"
)

// IndexItems slice type alias for IndexItem
type IndexItems []IndexItem

// IndexItem define an index item
type IndexItem struct {
	Unique bool
	Name   string
	Column string
	Order  int
	Sort   sortenum.Sort
}

// Len return length of IndexItems
func (it IndexItems) Len() int {
	return len(it)
}

// Less define asc or desc order
func (it IndexItems) Less(i, j int) bool {
	return it[i].Order < it[j].Order
}

// Swap change position of elements at i and j
func (it IndexItems) Swap(i, j int) {
	it[i], it[j] = it[j], it[i]
}

// Index define an index
type Index struct {
	Table  string
	Unique bool
	Name   string
	Items  []IndexItem
}

const indexsqltmpl = `{{define "drop"}}
ALTER TABLE ` + "`" + `{{.Table}}` + "`" + ` DROP INDEX ` + "`" + `{{.Name}}` + "`" + `;
{{end}}

{{define "add"}}
ALTER TABLE ` + "`" + `{{.Table}}` + "`" + ` ADD {{if .Unique}}UNIQUE{{end}} INDEX ` + "`" + `{{.Name}}` + "`" + ` ({{range $j, $it := .Items}}{{if $j}},{{end}}` + "`" + `{{$it.Column}}` + "`" + ` {{$it.Sort}}{{end}});
{{end}}`

func (idx *Index) DropIndexSql() (string, error) {
	return templateutils.StringBlock("index.tmpl", indexsqltmpl, "drop", idx)
}

func (idx *Index) AddIndexSql() (string, error) {
	return templateutils.StringBlock("index.tmpl", indexsqltmpl, "add", idx)
}

func NewIndexFromDbIndexes(dbIndexes []DbIndex) Index {
	unique := !dbIndexes[0].NonUnique
	idxName := dbIndexes[0].KeyName
	items := make([]IndexItem, len(dbIndexes))
	for i, idx := range dbIndexes {
		var sor sortenum.Sort
		if idx.Collation == "B" {
			sor = sortenum.Desc
		} else {
			sor = sortenum.Asc
		}
		items[i] = IndexItem{
			Column: idx.ColumnName,
			Order:  idx.SeqInIndex,
			Sort:   sor,
		}
	}
	it := IndexItems(items)
	sort.Stable(it)
	return Index{
		Unique: unique,
		Name:   idxName,
		Items:  it,
	}
}

func toColumnType(goType string) columnenum.ColumnType {
	switch goType {
	case "int", "int16", "int32":
		return columnenum.IntType
	case "int64":
		return columnenum.BigintType
	case "float32":
		return columnenum.FloatType
	case "float64":
		return columnenum.DoubleType
	case "string":
		return columnenum.VarcharType
	case "bool", "int8":
		return columnenum.TinyintType
	case "time.Time":
		return columnenum.DatetimeType
	case "decimal.Decimal":
		return "decimal(6,2)"
	case "types.JSONText":
		return columnenum.JSONType
	}
	panic(fmt.Sprintf("no available type %s", goType))
}

func toGoType(colType columnenum.ColumnType, nullable bool) string {
	var goType string
	if nullable {
		goType += "*"
	}
	if stringutils.HasPrefixI(string(colType), strings.ToLower(string(columnenum.IntType))) {
		goType += "int"
	} else if stringutils.HasPrefixI(string(colType), strings.ToLower(string(columnenum.BigintType))) {
		goType += "int64"
	} else if stringutils.HasPrefixI(string(colType), strings.ToLower(string(columnenum.FloatType))) {
		goType += "float32"
	} else if stringutils.HasPrefixI(string(colType), strings.ToLower(string(columnenum.DoubleType))) {
		goType += "float64"
	} else if stringutils.HasPrefixI(string(colType), "varchar") {
		goType += "string"
	} else if stringutils.HasPrefixI(string(colType), strings.ToLower(string(columnenum.TextType))) {
		goType += "string"
	} else if stringutils.HasPrefixI(string(colType), strings.ToLower(string(columnenum.TinyintType))) {
		goType += "int8"
	} else if stringutils.HasPrefixI(string(colType), strings.ToLower(string(columnenum.DatetimeType))) {
		goType += "time.Time"
	} else if stringutils.HasPrefixI(string(colType), strings.ToLower(string(columnenum.MediumtextType))) {
		goType += "string"
	} else if stringutils.HasPrefixI(string(colType), strings.ToLower(string(columnenum.DecimalType))) {
		goType += "decimal.Decimal"
	} else if stringutils.HasPrefixI(string(colType), strings.ToLower(string(columnenum.LongtextType))) {
		goType += "string"
	} else if stringutils.HasPrefixI(string(colType), strings.ToLower(string(columnenum.JSONType))) {
		goType += "types.JSONText"
	} else {
		panic(fmt.Sprintf("no available type %s", colType))
	}
	return goType
}

// CheckPk check key is primary key or not
func CheckPk(key keyenum.Key) bool {
	return key == keyenum.Pri
}

// CheckNull check a column is nullable or not
func CheckNull(null nullenum.Null) bool {
	return null == nullenum.Yes
}

// CheckUnsigned check a column is unsigned or not
func CheckUnsigned(dbColType string) bool {
	splits := strings.Split(dbColType, " ")
	if len(splits) == 1 {
		return false
	}
	return splits[1] == "unsigned"
}

// CheckAutoincrement check a column is auto increment or not
func CheckAutoincrement(extra string) bool {
	return strings.Contains(extra, "auto_increment")
}

// CheckAutoSet check a column is auto generated by database or not
func CheckAutoSet(defaultVal string) bool {
	return strings.ToLower(defaultVal) == strings.ToLower(now)
}

// Column define a column
type Column struct {
	Table         string
	Name          string
	Type          columnenum.ColumnType
	Default       string
	Pk            bool
	Nullable      bool
	Unsigned      bool
	Autoincrement bool
	Extra         extraenum.Extra
	Meta          astutils.FieldMeta
	AutoSet       bool
	Indexes       []IndexItem
	Fk            ForeignKey
}

var altersqltmpl = `{{define "change"}}
ALTER TABLE ` + "`" + `{{.Table}}` + "`" + `
CHANGE COLUMN ` + "`" + `{{.Name}}` + "`" + ` ` + "`" + `{{.Name}}` + "`" + ` {{.Type}} {{if .Nullable}}NULL{{else}}NOT NULL{{end}}{{if .Autoincrement}} AUTO_INCREMENT{{end}}{{if .Default}} DEFAULT {{.Default}}{{end}}{{if .Extra}} {{.Extra}}{{end}};
{{end}}

{{define "add"}}
ALTER TABLE ` + "`" + `{{.Table}}` + "`" + `
ADD COLUMN ` + "`" + `{{.Name}}` + "`" + ` {{.Type}} {{if .Nullable}}NULL{{else}}NOT NULL{{end}}{{if .Autoincrement}} AUTO_INCREMENT{{end}}{{if .Default}} DEFAULT {{.Default}}{{end}}{{if .Extra}} {{.Extra}}{{end}};
{{end}}
`

// ChangeColumnSql return change column sql
func (c *Column) ChangeColumnSql() (string, error) {
	return templateutils.StringBlock("alter.tmpl", altersqltmpl, "change", c)
}

// AddColumnSql return add column sql
func (c *Column) AddColumnSql() (string, error) {
	return templateutils.StringBlock("alter.tmpl", altersqltmpl, "add", c)
}

// DbColumn defines a column
type DbColumn struct {
	Field   string        `db:"Field"`
	Type    string        `db:"Type"`
	Null    nullenum.Null `db:"Null"`
	Key     keyenum.Key   `db:"Key"`
	Default *string       `db:"Default"`
	Extra   string        `db:"Extra"`
	Comment string        `db:"Comment"`
}

// DbIndex defines an index refer to https://www.mysqltutorial.org/mysql-index/mysql-show-indexes/
type DbIndex struct {
	Table      string `db:"Table"`        // The name of the table
	NonUnique  bool   `db:"Non_unique"`   // 1 if the index can contain duplicates, 0 if it cannot.
	KeyName    string `db:"Key_name"`     // The name of the index. The primary key index always has the name of PRIMARY.
	SeqInIndex int    `db:"Seq_in_index"` // The column sequence number in the index. The first column sequence number starts from 1.
	ColumnName string `db:"Column_name"`  // The column name
	Collation  string `db:"Collation"`    // Collation represents how the column is sorted in the index. A means ascending, B means descending, or NULL means not sorted.
}

// DbForeignKey from INFORMATION_SCHEMA.KEY_COLUMN_USAGE
type DbForeignKey struct {
	TableName            string `db:"TABLE_NAME"`
	ColumnName           string `db:"COLUMN_NAME"`
	ConstraintName       string `db:"CONSTRAINT_NAME"`
	ReferencedTableName  string `db:"REFERENCED_TABLE_NAME"`
	ReferencedColumnName string `db:"REFERENCED_COLUMN_NAME"`
}

// DbAction from information_schema.REFERENTIAL_CONSTRAINTS
type DbAction struct {
	TableName           string `db:"TABLE_NAME"`
	ConstraintName      string `db:"CONSTRAINT_NAME"`
	ReferencedTableName string `db:"REFERENCED_TABLE_NAME"`
	UpdateRule          string `db:"UPDATE_RULE"`
	DeleteRule          string `db:"DELETE_RULE"`
}

type ForeignKey struct {
	// Table the child table
	Table string
	// Constraint name of foreign key constraint
	Constraint string
	// Fk foreign key
	Fk string
	// ReferencedTable the referenced table
	ReferencedTable string
	// ReferencedCol the referenced column of ReferencedTable
	ReferencedCol string
	UpdateRule    string
	DeleteRule    string
	FullRule      string
}

const fksqltmpl = `{{define "drop"}}
ALTER TABLE ` + "`" + `{{.Table}}` + "`" + ` DROP FOREIGN KEY {{.Constraint}};
{{end}}

{{define "add"}}
ALTER TABLE ` + "`" + `{{.Table}}` + "`" + ` ADD CONSTRAINT {{.Constraint}} FOREIGN KEY ({{.Fk}}) REFERENCES {{.ReferencedTable}}({{.ReferencedCol}}) {{.FullRule}};
{{end}}`

func (fk *ForeignKey) DropFkSql() (string, error) {
	return templateutils.StringBlock("fk.tmpl", fksqltmpl, "drop", fk)
}

func (fk *ForeignKey) AddFkSql() (string, error) {
	return templateutils.StringBlock("fk.tmpl", fksqltmpl, "add", fk)
}

// Table defines a table
type Table struct {
	Name    string
	Columns []Column
	Pk      string
	Indexes []Index
	Meta    astutils.StructMeta
	Fks     []ForeignKey
}

// NewTableFromStruct creates a Table instance from structMeta
func NewTableFromStruct(structMeta astutils.StructMeta, prefix ...string) Table {
	var (
		columns       []Column
		uniqueindexes []Index
		indexes       []Index
		fks           []ForeignKey
		pkColumn      Column
		table         string
	)
	table = strcase.ToSnake(structMeta.Name)
	if len(prefix) > 0 {
		table = prefix[0] + table
	}
	for _, field := range structMeta.Fields {
		var (
			columnName     string
			_uniqueindexes []Index
			_indexes       []Index
			_fks           []ForeignKey
			column         Column
		)
		column.Table = table
		column.Meta = field
		columnName = strcase.ToSnake(field.Name)
		column.Name = columnName
		if stringutils.IsNotEmpty(field.Tag) {
			tags := strings.Split(field.Tag, `" `)
			var ddTag string
			for _, tag := range tags {
				if strings.HasPrefix(tag, "dd:") {
					ddTag = strings.Trim(strings.TrimPrefix(tag, "dd:"), `"`)
					break
				}
			}
			if stringutils.IsNotEmpty(ddTag) {
				_indexes, _uniqueindexes, _fks = parseDdTag(ddTag, field, &column)
			}
		}

		if strings.HasPrefix(field.Type, "*") {
			column.Nullable = true
		}

		if stringutils.IsEmpty(string(column.Type)) {
			column.Type = toColumnType(strings.TrimPrefix(field.Type, "*"))
		}

		for _, idx := range _indexes {
			if stringutils.IsNotEmpty(idx.Name) {
				idx.Items[0].Column = columnName
				indexes = append(indexes, idx)
			}
		}

		for _, uidx := range _uniqueindexes {
			if stringutils.IsNotEmpty(uidx.Name) {
				uidx.Items[0].Column = columnName
				uniqueindexes = append(uniqueindexes, uidx)
			}
		}

		for _, fk := range _fks {
			if stringutils.IsNotEmpty(fk.Fk) {
				fks = append(fks, fk)
			}
		}

		columns = append(columns, column)
	}

	for _, column := range columns {
		if column.Pk {
			pkColumn = column
			break
		}
	}

	indexesResult := mergeIndexes(indexes, uniqueindexes)

	return Table{
		Name:    table,
		Columns: columns,
		Pk:      pkColumn.Name,
		Indexes: indexesResult,
		Meta:    structMeta,
		Fks:     fks,
	}
}

type sortableIndexes []Index

// Len return length of sortableIndexes
func (it sortableIndexes) Len() int {
	return len(it)
}

// Less define asc or desc order
func (it sortableIndexes) Less(i, j int) bool {
	return it[i].Name < it[j].Name
}

// Swap change position of elements at i and j
func (it sortableIndexes) Swap(i, j int) {
	it[i], it[j] = it[j], it[i]
}

func mergeIndexes(indexes, uniqueindexes []Index) []Index {
	uniqueMap := make(map[string][]IndexItem)
	indexMap := make(map[string][]IndexItem)

	for _, unique := range uniqueindexes {
		if items, exists := uniqueMap[unique.Name]; exists {
			items = append(items, unique.Items...)
			uniqueMap[unique.Name] = items
		} else {
			uniqueMap[unique.Name] = unique.Items
		}
	}

	for _, index := range indexes {
		if items, exists := indexMap[index.Name]; exists {
			items = append(items, index.Items...)
			indexMap[index.Name] = items
		} else {
			indexMap[index.Name] = index.Items
		}
	}

	var uniquesResult, indexesResult []Index

	for k, v := range uniqueMap {
		it := IndexItems(v)
		sort.Stable(it)
		uniquesResult = append(uniquesResult, Index{
			Unique: true,
			Name:   k,
			Items:  it,
		})
	}

	for k, v := range indexMap {
		it := IndexItems(v)
		sort.Stable(it)
		indexesResult = append(indexesResult, Index{
			Name:  k,
			Items: it,
		})
	}

	sort.Stable(sortableIndexes(indexesResult))
	sort.Stable(sortableIndexes(uniquesResult))

	indexesResult = append(indexesResult, uniquesResult...)
	return indexesResult
}

func parseDdTag(ddTag string, field astutils.FieldMeta, column *Column) (indexes []Index, uniqueIndexes []Index, fks []ForeignKey) {
	kvs := strings.Split(ddTag, ";")
	for _, kv := range kvs {
		pair := strings.Split(kv, ":")
		if len(pair) > 1 {
			parsePair(pair, column, &indexes, &uniqueIndexes, &fks)
		} else {
			parseSingle(pair, column, field, &indexes, &uniqueIndexes)
		}
	}
	return
}

func parseSingle(pair []string, column *Column, field astutils.FieldMeta, indexes *[]Index, uniqueIndexes *[]Index) {
	key := pair[0]
	switch key {
	case "pk":
		column.Pk = true
		break
	case "null":
		column.Nullable = true
		break
	case "unsigned":
		column.Unsigned = true
		break
	case "auto":
		column.Autoincrement = true
		break
	case "index":
		*indexes = append(*indexes, Index{
			Name: strcase.ToSnake(field.Name) + "_idx",
			Items: []IndexItem{
				{
					Order: 1,
					Sort:  sortenum.Asc,
				},
			},
		})
		break
	case "unique":
		*uniqueIndexes = append(*uniqueIndexes, Index{
			Name: strcase.ToSnake(field.Name) + "_idx",
			Items: []IndexItem{
				{
					Order: 1,
					Sort:  sortenum.Asc,
				},
			},
		})
		break
	}
}

func parsePair(pair []string, column *Column, indexes *[]Index, uniqueIndexes *[]Index, fks *[]ForeignKey) {
	key := pair[0]
	value := pair[1]
	if stringutils.IsEmpty(value) {
		panic(fmt.Sprintf("%+v", errors.New("value should not be empty")))
	}
	switch key {
	case "type":
		column.Type = columnenum.ColumnType(value)
		break
	case "default":
		column.Default = value
		column.AutoSet = CheckAutoSet(value)
		break
	case "extra":
		column.Extra = extraenum.Extra(value)
		break
	case "index":
		props := strings.Split(value, ",")
		indexName := props[0]
		order := props[1]
		orderInt, err := strconv.Atoi(order)
		if err != nil {
			panic(err)
		}
		var sor sortenum.Sort
		if len(props) < 3 || stringutils.IsEmpty(props[2]) {
			sor = sortenum.Asc
		} else {
			sor = sortenum.Sort(props[2])
		}
		*indexes = append(*indexes, Index{
			Name: indexName,
			Items: []IndexItem{
				{
					Order: orderInt,
					Sort:  sor,
				},
			},
		})
		break
	case "unique":
		props := strings.Split(value, ",")
		indexName := props[0]
		order := props[1]
		orderInt, err := strconv.Atoi(order)
		if err != nil {
			panic(err)
		}
		var sort sortenum.Sort
		if len(props) < 3 || stringutils.IsEmpty(props[2]) {
			sort = sortenum.Asc
		} else {
			sort = sortenum.Sort(props[2])
		}
		*uniqueIndexes = append(*uniqueIndexes, Index{
			Name: indexName,
			Items: []IndexItem{
				{
					Order: orderInt,
					Sort:  sort,
				},
			},
		})
		break
	case "fk":
		props := strings.Split(value, ",")
		refTable := props[0]
		var refCol string
		if len(props) > 1 {
			refCol = props[1]
		} else {
			refCol = "id"
		}
		var constraint string
		if len(props) > 2 {
			constraint = props[2]
		} else {
			constraint = fmt.Sprintf("fk_%s_%s_%s", column.Name, refTable, refCol)
		}
		var fullRule string
		if len(props) > 3 {
			fullRule = props[3]
		}
		*fks = append(*fks, ForeignKey{
			Table:           column.Table,
			Constraint:      constraint,
			Fk:              column.Name,
			ReferencedTable: refTable,
			ReferencedCol:   refCol,
			FullRule:        fullRule,
		})
		break
	}
}

// NewFieldFromColumn creates an astutils.FieldMeta instance from col
func NewFieldFromColumn(col Column) astutils.FieldMeta {
	tag := "dd:"
	var feats []string
	if col.Pk {
		feats = append(feats, "pk")
	}
	if col.Autoincrement {
		feats = append(feats, "auto")
	}
	goType := toGoType(col.Type, col.Nullable)
	if col.Nullable && !strings.HasPrefix(goType, "*") {
		feats = append(feats, "null")
	}
	if stringutils.IsNotEmpty(string(col.Type)) {
		feats = append(feats, fmt.Sprintf("type:%s", string(col.Type)))
	}
	if stringutils.IsNotEmpty(col.Default) {
		val := col.Default
		re := regexp.MustCompile(`^\(.+\)$`)
		var defaultClause string
		if strings.ToUpper(val) == "CURRENT_TIMESTAMP" || re.MatchString(val) {
			defaultClause = fmt.Sprintf("default:%s", val)
		} else {
			defaultClause = fmt.Sprintf("default:'%s'", val)
		}
		feats = append(feats, defaultClause)
	}
	if stringutils.IsNotEmpty(string(col.Extra)) {
		feats = append(feats, fmt.Sprintf("extra:%s", string(col.Extra)))
	}
	for _, idx := range col.Indexes {
		var indexClause string
		if idx.Name == "PRIMARY" {
			continue
		}
		if idx.Unique {
			indexClause = "unique:"
		} else {
			indexClause = "index:"
		}
		indexClause += fmt.Sprintf("%s,%d,%s", idx.Name, idx.Order, string(idx.Sort))
		feats = append(feats, indexClause)
	}
	if stringutils.IsNotEmpty(col.Fk.Constraint) {
		feats = append(feats, fmt.Sprintf("fk:%s,%s,%s,%s", col.Fk.ReferencedTable, col.Fk.ReferencedCol, col.Fk.Constraint, col.Fk.FullRule))
	}
	return astutils.FieldMeta{
		Name: strcase.ToCamel(col.Name),
		Type: goType,
		Tag:  fmt.Sprintf(`%s"%s"`, tag, strings.Join(feats, ";")),
	}
}

var createsqltmpl = `CREATE TABLE ` + "`" + `{{.Name}}` + "`" + ` (
{{- range $co := .Columns }}
` + "`" + `{{$co.Name}}` + "`" + ` {{$co.Type}} {{if $co.Nullable}}NULL{{else}}NOT NULL{{end}}{{if $co.Autoincrement}} AUTO_INCREMENT{{end}}{{if $co.Default}} DEFAULT {{$co.Default}}{{end}}{{if $co.Extra}} {{$co.Extra}}{{end}},
{{- end }}
PRIMARY KEY (` + "`" + `{{.Pk}}` + "`" + `){{if .Indexes}},{{end}}
{{- range $i, $ind := .Indexes}}
{{- if $i}},{{end}}
{{if $ind.Unique}}UNIQUE {{end}}INDEX ` + "`" + `{{$ind.Name}}` + "`" + ` ({{ range $j, $it := $ind.Items }}{{if $j}},{{end}}` + "`" + `{{$it.Column}}` + "`" + ` {{$it.Sort}}{{ end }})
{{- end }}{{if .Fks}},{{end}}
{{- range $i, $fk := .Fks}}
{{- if $i}},{{end}}
CONSTRAINT ` + "`" + `{{$fk.Constraint}}` + "`" + ` FOREIGN KEY (` + "`" + `{{$fk.Fk}}` + "`" + `)
REFERENCES ` + "`" + `{{$fk.ReferencedTable}}` + "`" + `(` + "`" + `{{$fk.ReferencedCol}}` + "`" + `)
{{- if $fk.FullRule}}
{{$fk.FullRule}}
{{- end }}
{{- end }})`

// CreateSql return create table sql
func (t *Table) CreateSql() (string, error) {
	return templateutils.String("create.sql.tmpl", createsqltmpl, t)
}
