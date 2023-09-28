package main

//note many features are not supported. Only most basic iota syntax can be translated

import (
	rs "github.com/cshu/golangrs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var constRe = regexp.MustCompile(`^const\s*\(\s*(//.*)?$`)
var offsetRe = regexp.MustCompile(`([_A-Za-z0-9]\s+([_A-Za-z0-9]+)\s*)?=\s*iota\s*(\+\s*([0-9]+)\s*)?(//.*)?$`)
var identifierRe = regexp.MustCompile(`^\s*([_A-Za-z0-9]+)`)
var constEndRe = regexp.MustCompile(`^\s*\)\s*(//.*)?$`)
var sb strings.Builder

const generatorVersion = `enum-translator version 20230928.1`

func init() {
}

var filenmExt string
var filenmExcludeExt string
var typeStr string
var offset int //todo use int64?
func writeLine(line string) {
	identifier := identifierRe.FindAllStringSubmatch(line, -1)[0][1]
	if `_` != identifier {
		outputLangUtil.WriteForIdentifier(identifier)
	}
	offset++
}

type OutputLangUtil interface {
	WriteForIdentifier(string)
	WriteHeader()
	WriteFooter()
}
type OutputJsUtil struct {
}
type OutputJavaUtil struct {
}

func (util *OutputJsUtil) WriteHeader() {
	sb.WriteString(useStrict + "//This is GENERATED code!! (" + generatorVersion + ") Do NOT modify manually!!\n")
}
func (util *OutputJsUtil) WriteFooter() {
}
func (util *OutputJavaUtil) WriteHeader() {
	sb.WriteString("//This is GENERATED code!! (" + generatorVersion + ") Do NOT modify manually!!\n")
	sb.WriteString("public class " + filenmExcludeExt + " {\n")
}
func (util *OutputJavaUtil) WriteFooter() {
	sb.WriteString("}\n")
}
func (util *OutputJsUtil) WriteForIdentifier(identifier string) {
	sb.WriteString(`const ` + identifier + ` = ` + strconv.Itoa(offset) + `;`)
	sb.WriteByte('\n')
}
func (util *OutputJavaUtil) WriteForIdentifier(identifier string) {
	if `byte` == typeStr {
		sb.WriteString("\tpublic static final byte " + identifier + ` = `)
		if offset > 127 {
			sb.WriteString(`(byte)`)
		}
		sb.WriteString(strconv.Itoa(offset) + `;`)
	} else {
		sb.WriteString("\tpublic static final int " + identifier + ` = ` + strconv.Itoa(offset) + `;`)
	}
	sb.WriteByte('\n')
}

var useStrict string = "'use strict';\n\n"
var outputLangUtil OutputLangUtil

func main() {
	var err error
	if 2 != len(os.Args) {
		println(`Please provide 1 file as argument`)
		return
	}
	outfilename := os.Getenv("OUTFILENAME")
	if "" == outfilename {
		println(`Env variable OUTFILENAME not set`)
		return
	}
	jsUseStrict := os.Getenv("JS_USE_STRICT")
	if "" == jsUseStrict {
		useStrict = ""
	}
	filepathBase := filepath.Base(outfilename)
	filenmExt = filepath.Ext(filepathBase)
	switch filenmExt {
	case `.js`:
		outputLangUtil = &OutputJsUtil{}
	case `.java`:
		outputLangUtil = &OutputJavaUtil{}
	default:
		println(`OUTFILENAME extension not supported`)
		return
	}
	filenmExcludeExt = filepathBase[:len(filepathBase)-len(filenmExt)]
	outputLangUtil.WriteHeader()
	var isLineConst bool
	var enumFound bool
	rs.ReadHugeFilelines(os.Args[1], func(line string) bool {
		lastLineConst := isLineConst
		isLineConst = false
		if lastLineConst {
			offsetMatches := offsetRe.FindAllStringSubmatch(line, -1)
			if len(offsetMatches) != 0 {
				enumFound = true
				//println(line)
				rs.AssertElsePanicWithStr(6 == len(offsetMatches[0]), `This should never happen!`)
				//if len(offsetMatches[0])>=3{
				offStr := offsetMatches[0][4]
				if "" == offStr {
					offset = 0
				} else {
					offset, err = strconv.Atoi(offStr)
					rs.CheckErr(err)
				}
				typeStr = offsetMatches[0][2]
				//}else{
				//	offset = 0
				//}
				sb.WriteByte('\n')
				writeLine(line)
			}
			return true
		}
		if enumFound {
			if constEndRe.MatchString(line) {
				enumFound = false
			} else {
				writeLine(line)
			}
			return true
		}
		isLineConst = constRe.MatchString(line)
		return true
	})
	outputLangUtil.WriteFooter()
	err = ioutil.WriteFile(outfilename, []byte(sb.String()), 0644)
	rs.CheckErr(err)
	println(`Done`)
}
