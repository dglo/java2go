package grammar

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

//%e 1600
//%n 800
//%p 5000

type ByteReader interface {
	ReadByte() (byte, error)
}

type myLexer struct{
    path    string
    fd      *os.File

    debugLex bool

    src     ByteReader
    buf     []byte
    empty   bool

    data    []byte

    linenum int
    col     int
    current byte

    program *JProgramFile
}

func NewFileLexer(path string, debugLex bool) (y *myLexer) {
    fd, err := os.Open(path)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Cannot open \"%s\": %v\n", path, err)
        return nil
    }

    y = NewLexer(bufio.NewReader(fd), debugLex)
    y.path = path
    y.fd = fd
    return
}

func NewLexer(src ByteReader, debugLex bool) (y *myLexer) {
    y = &myLexer{src: src, debugLex: debugLex}
    if b, err := src.ReadByte(); err == nil {
        y.current = b
        y.linenum = 1
        y.col = 1
    }
    return
}

func (y *myLexer) clearbuf() {
    y.buf = y.buf[:0]
}

func (y *myLexer) getc() byte {
    if y.current != 0 {
        y.buf = append(y.buf, y.current)
    }

    var b byte
    var err error

    if b, err = y.src.ReadByte(); err != nil {
        y.current = 0
        return 0
    }

    y.current = b
    if b != '\n' {
        y.col++
        y.data = append(y.data, b)
    } else {
        y.linenum++
        y.col = 1
        y.data = y.data[:0]
    }

    if y.debugLex {
        fmt.Fprintf(os.Stderr, "getc -> %c (l%d c%d)\n", y.current, y.linenum,
                    y.col)
    }

    return y.current
}

func (y myLexer) Close() {
    if y.fd != nil {
        y.fd.Close()
    }
}

func (y myLexer) Error(e string) {
    var front string
    if y.path != "" {
        front = fmt.Sprintf("\"%s\" ", y.path)
    }

    log.Printf("At %sline %d, column %d: %s\n%s\n", front, y.linenum, y.col,
	       e, string(y.data))
}

func (y myLexer) LexChar(lval *JulySymType) int {
    lval.str = string(y.buf)
    lval.obj = nil
    if y.debugLex {
        fmt.Printf("LexChar -> '%c'\n", y.buf[0])
    }
    return int(y.buf[0])
}

func (y myLexer) LexString(token int, lval *JulySymType) int {
    lval.str = string(y.buf)
    lval.token = token
    lval.obj = nil
    if y.debugLex {
        fmt.Printf("LexString -> tok-%d \"%s\"\n", token, y.buf)
    }
    return token
}

func (y *myLexer) JavaProgram() *JProgramFile {
    return y.program
}

func (y *myLexer) SetJavaProgram(prog *JProgramFile) {
    y.program = prog
}

func (y *myLexer) String() string {
    return fmt.Sprintf("myLexer[%s data \"%s\" num %d col %d buf \"%s\"" +
		       " current '%c']", y.path, string(y.data), y.linenum,
		       y.col, y.buf, y.current)
}

func (y *myLexer) Lex(lval *JulySymType) int {
    c := y.current
    if y.empty {
        c, y.empty = y.getc(), false
    }



yystate0:

        y.clearbuf()
        // clear all cached values
        lval.str = ""
        lval.name = nil
        lval.namelist = nil
        lval.obj = nil
        lval.objlist = nil
        lval.count = 0

goto yystart1

goto yystate1 // silence unused label error
yystate1:
c = y.getc()
yystart1:
switch {
default:
goto yyabort
case c == ' ':
goto yystate8
case c == '!':
goto yystate9
case c == '"':
goto yystate11
case c == '%':
goto yystate14
case c == '&':
goto yystate16
case c == '(' || c == ')' || c == ',' || c == ';' || c == ']' || c == '{' || c == '}':
goto yystate24
case c == '*':
goto yystate25
case c == '+':
goto yystate27
case c == '-':
goto yystate30
case c == '.':
goto yystate33
case c == '/':
goto yystate38
case c == '0':
goto yystate46
case c == ':' || c == '<' || c >= '>' && c <= '@' || c == '~':
goto yystate52
case c == '=':
goto yystate53
case c == '[':
goto yystate55
case c == '\'':
goto yystate19
case c == '\b':
goto yystate3
case c == '\f':
goto yystate6
case c == '\n':
goto yystate5
case c == '\r':
goto yystate7
case c == '\t':
goto yystate4
case c == '\x01' || c >= 'A' && c <= 'Z' || c == '_' || c == 'g' || c == 'h' || c == 'j' || c == 'k' || c == 'm' || c == 'o' || c == 'q' || c == 'u' || c >= 'x' && c <= 'z':
goto yystate2
case c == '^':
goto yystate76
case c == 'a':
goto yystate78
case c == 'b':
goto yystate91
case c == 'c':
goto yystate105
case c == 'd':
goto yystate126
case c == 'e':
goto yystate138
case c == 'f':
goto yystate151
case c == 'i':
goto yystate168
case c == 'l':
goto yystate198
case c == 'n':
goto yystate202
case c == 'p':
goto yystate213
case c == 'r':
goto yystate238
case c == 's':
goto yystate244
case c == 't':
goto yystate274
case c == 'v':
goto yystate293
case c == 'w':
goto yystate303
case c == '|':
goto yystate308
case c >= '1' && c <= '9':
goto yystate51
}

yystate2:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate3:
c = y.getc()
goto yyrule74

yystate4:
c = y.getc()
goto yyrule73

yystate5:
c = y.getc()
goto yyrule71

yystate6:
c = y.getc()
goto yyrule72

yystate7:
c = y.getc()
goto yyrule70

yystate8:
c = y.getc()
goto yyrule75

yystate9:
c = y.getc()
switch {
default:
goto yyrule4
case c == '=':
goto yystate10
}

yystate10:
c = y.getc()
goto yyrule7

yystate11:
c = y.getc()
switch {
default:
goto yyabort
case c == '"':
goto yystate12
case c == '\\':
goto yystate13
case c >= '\x01' && c <= '!' || c >= '#' && c <= '[' || c >= ']' && c <= 'ÿ':
goto yystate11
}

yystate12:
c = y.getc()
goto yyrule69

yystate13:
c = y.getc()
switch {
default:
goto yyabort
case c == '"' || c == '\'' || c >= '0' && c <= '7' || c == '\\' || c == 'b' || c == 'f' || c == 'n' || c == 'r' || c == 't' || c == 'u':
goto yystate11
}

yystate14:
c = y.getc()
switch {
default:
goto yyrule4
case c == '=':
goto yystate15
}

yystate15:
c = y.getc()
goto yyrule19

yystate16:
c = y.getc()
switch {
default:
goto yyrule4
case c == '&':
goto yystate17
case c == '=':
goto yystate18
}

yystate17:
c = y.getc()
goto yyrule9

yystate18:
c = y.getc()
goto yyrule16

yystate19:
c = y.getc()
switch {
default:
goto yyabort
case c == '\\':
goto yystate21
case c >= '\x01' && c <= '&' || c >= '(' && c <= '[' || c >= ']' && c <= 'ÿ':
goto yystate20
}

yystate20:
c = y.getc()
switch {
default:
goto yyabort
case c == '\'':
goto yystate12
}

yystate21:
c = y.getc()
switch {
default:
goto yyabort
case c == '"' || c == '\'' || c == '\\' || c == 'b' || c == 'f' || c == 'n' || c == 'r' || c == 't' || c == 'u':
goto yystate20
case c >= '0' && c <= '3':
goto yystate22
case c >= '4' && c <= '7':
goto yystate23
}

yystate22:
c = y.getc()
switch {
default:
goto yyabort
case c == '\'':
goto yystate12
case c >= '0' && c <= '7':
goto yystate23
}

yystate23:
c = y.getc()
switch {
default:
goto yyabort
case c == '\'':
goto yystate12
case c >= '0' && c <= '7':
goto yystate20
}

yystate24:
c = y.getc()
goto yyrule3

yystate25:
c = y.getc()
switch {
default:
goto yyrule4
case c == '=':
goto yystate26
}

yystate26:
c = y.getc()
goto yyrule14

yystate27:
c = y.getc()
switch {
default:
goto yyrule4
case c == '+':
goto yystate28
case c == '=':
goto yystate29
}

yystate28:
c = y.getc()
goto yyrule10

yystate29:
c = y.getc()
goto yyrule12

yystate30:
c = y.getc()
switch {
default:
goto yyrule4
case c == '-':
goto yystate31
case c == '=':
goto yystate32
}

yystate31:
c = y.getc()
goto yyrule11

yystate32:
c = y.getc()
goto yyrule13

yystate33:
c = y.getc()
switch {
default:
goto yyrule3
case c >= '0' && c <= '9':
goto yystate34
}

yystate34:
c = y.getc()
switch {
default:
goto yyrule69
case c == 'D' || c == 'F' || c == 'd' || c == 'f':
goto yystate12
case c == 'E' || c == 'e':
goto yystate35
case c >= '0' && c <= '9':
goto yystate34
}

yystate35:
c = y.getc()
switch {
default:
goto yyrule69
case c == '+' || c == '-':
goto yystate36
case c == 'D' || c == 'F' || c == 'd' || c == 'f':
goto yystate12
case c >= '0' && c <= '9':
goto yystate37
}

yystate36:
c = y.getc()
switch {
default:
goto yyabort
case c >= '0' && c <= '9':
goto yystate37
}

yystate37:
c = y.getc()
switch {
default:
goto yyrule69
case c == 'D' || c == 'F' || c == 'd' || c == 'f':
goto yystate12
case c >= '0' && c <= '9':
goto yystate37
}

yystate38:
c = y.getc()
switch {
default:
goto yyrule4
case c == '*':
goto yystate39
case c == '/':
goto yystate44
case c == '=':
goto yystate45
}

yystate39:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate41
case c == '/':
goto yystate43
case c >= '\x01' && c <= ')' || c >= '+' && c <= '.' || c >= '0' && c <= 'ÿ':
goto yystate40
}

yystate40:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate41
case c >= '\x01' && c <= ')' || c >= '+' && c <= 'ÿ':
goto yystate40
}

yystate41:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate41
case c == '/':
goto yystate42
case c >= '\x01' && c <= ')' || c >= '+' && c <= '.' || c >= '0' && c <= 'ÿ':
goto yystate40
}

yystate42:
c = y.getc()
switch {
default:
goto yyrule76
case c == '/':
goto yystate39
}

yystate43:
c = y.getc()
switch {
default:
goto yyabort
case c == '/':
goto yystate39
}

yystate44:
c = y.getc()
switch {
default:
goto yyrule76
case c >= '\x01' && c <= '\t' || c >= '\v' && c <= 'ÿ':
goto yystate44
}

yystate45:
c = y.getc()
goto yyrule15

yystate46:
c = y.getc()
switch {
default:
goto yyrule69
case c == '.':
goto yystate34
case c == '8' || c == '9':
goto yystate48
case c == 'D' || c == 'F' || c == 'L' || c == 'd' || c == 'f' || c == 'l':
goto yystate12
case c == 'E' || c == 'e':
goto yystate35
case c == 'X' || c == 'x':
goto yystate49
case c >= '0' && c <= '7':
goto yystate47
}

yystate47:
c = y.getc()
switch {
default:
goto yyrule69
case c == '.':
goto yystate34
case c == '8' || c == '9':
goto yystate48
case c == 'D' || c == 'F' || c == 'L' || c == 'd' || c == 'f' || c == 'l':
goto yystate12
case c == 'E' || c == 'e':
goto yystate35
case c >= '0' && c <= '7':
goto yystate47
}

yystate48:
c = y.getc()
switch {
default:
goto yyabort
case c == '.':
goto yystate34
case c == 'D' || c == 'F' || c == 'd' || c == 'f':
goto yystate12
case c == 'E' || c == 'e':
goto yystate35
case c >= '0' && c <= '9':
goto yystate48
}

yystate49:
c = y.getc()
switch {
default:
goto yyabort
case c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'a' && c <= 'f':
goto yystate50
}

yystate50:
c = y.getc()
switch {
default:
goto yyrule69
case c == 'L' || c == 'l':
goto yystate12
case c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'a' && c <= 'f':
goto yystate50
}

yystate51:
c = y.getc()
switch {
default:
goto yyrule69
case c == '.':
goto yystate34
case c == 'D' || c == 'F' || c == 'L' || c == 'd' || c == 'f' || c == 'l':
goto yystate12
case c == 'E' || c == 'e':
goto yystate35
case c >= '0' && c <= '9':
goto yystate51
}

yystate52:
c = y.getc()
goto yyrule4

yystate53:
c = y.getc()
switch {
default:
goto yyrule4
case c == '=':
goto yystate54
}

yystate54:
c = y.getc()
goto yyrule6

yystate55:
c = y.getc()
switch {
default:
goto yyrule3
case c == '/':
goto yystate57
case c == ']':
goto yystate74
case c >= '\b' && c <= '\n' || c == '\f' || c == '\r' || c == ' ':
goto yystate56
}

yystate56:
c = y.getc()
switch {
default:
goto yyabort
case c == '/':
goto yystate57
case c == ']':
goto yystate74
case c >= '\b' && c <= '\n' || c == '\f' || c == '\r' || c == ' ':
goto yystate56
}

yystate57:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate58
case c == '/':
goto yystate64
}

yystate58:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate60
case c == '/':
goto yystate75
case c >= '\x01' && c <= ')' || c >= '+' && c <= '.' || c >= '0' && c <= 'ÿ':
goto yystate59
}

yystate59:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate60
case c >= '\x01' && c <= ')' || c >= '+' && c <= 'ÿ':
goto yystate59
}

yystate60:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate60
case c == '/':
goto yystate61
case c >= '\x01' && c <= ')' || c >= '+' && c <= '.' || c >= '0' && c <= 'ÿ':
goto yystate59
}

yystate61:
c = y.getc()
switch {
default:
goto yyabort
case c == '/':
goto yystate62
case c == ']':
goto yystate74
case c >= '\b' && c <= '\n' || c == '\f' || c == '\r' || c == ' ':
goto yystate56
}

yystate62:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate60
case c == '/':
goto yystate63
case c >= '\x01' && c <= ')' || c >= '+' && c <= '.' || c >= '0' && c <= 'ÿ':
goto yystate59
}

yystate63:
c = y.getc()
switch {
default:
goto yyabort
case c == '/':
goto yystate66
case c == '\n':
goto yystate56
case c == ']':
goto yystate73
case c >= '\x01' && c <= '\t' || c >= '\v' && c <= '.' || c >= '0' && c <= '\\' || c >= '^' && c <= 'ÿ':
goto yystate64
}

yystate64:
c = y.getc()
switch {
default:
goto yyabort
case c == '/':
goto yystate65
case c == '\n':
goto yystate56
case c == ']':
goto yystate73
case c >= '\x01' && c <= '\t' || c >= '\v' && c <= '.' || c >= '0' && c <= '\\' || c >= '^' && c <= 'ÿ':
goto yystate64
}

yystate65:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate66
case c == '/':
goto yystate65
case c == '\n':
goto yystate56
case c == ']':
goto yystate73
case c >= '\x01' && c <= '\t' || c >= '\v' && c <= ')' || c >= '+' && c <= '.' || c >= '0' && c <= '\\' || c >= '^' && c <= 'ÿ':
goto yystate64
}

yystate66:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate66
case c == '/':
goto yystate72
case c == '\n':
goto yystate68
case c == ']':
goto yystate71
case c >= '\x01' && c <= '\t' || c >= '\v' && c <= ')' || c >= '+' && c <= '.' || c >= '0' && c <= '\\' || c >= '^' && c <= 'ÿ':
goto yystate67
}

yystate67:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate66
case c == '\n':
goto yystate68
case c == ']':
goto yystate71
case c >= '\x01' && c <= '\t' || c >= '\v' && c <= ')' || c >= '+' && c <= '\\' || c >= '^' && c <= 'ÿ':
goto yystate67
}

yystate68:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate60
case c == '/':
goto yystate69
case c == ']':
goto yystate70
case c >= '\b' && c <= '\n' || c == '\f' || c == '\r' || c == ' ':
goto yystate68
case c >= '\x01' && c <= '\a' || c == '\v' || c >= '\x0e' && c <= '\x1f' || c >= '!' && c <= ')' || c >= '+' && c <= '.' || c >= '0' && c <= '\\' || c >= '^' && c <= 'ÿ':
goto yystate59
}

yystate69:
c = y.getc()
switch {
default:
goto yyabort
case c == '*':
goto yystate60
case c == '/':
goto yystate67
case c >= '\x01' && c <= ')' || c >= '+' && c <= '.' || c >= '0' && c <= 'ÿ':
goto yystate59
}

yystate70:
c = y.getc()
switch {
default:
goto yyrule5
case c == '*':
goto yystate60
case c >= '\x01' && c <= ')' || c >= '+' && c <= 'ÿ':
goto yystate59
}

yystate71:
c = y.getc()
switch {
default:
goto yyrule5
case c == '*':
goto yystate66
case c == '\n':
goto yystate68
case c == ']':
goto yystate71
case c >= '\x01' && c <= '\t' || c >= '\v' && c <= ')' || c >= '+' && c <= '\\' || c >= '^' && c <= 'ÿ':
goto yystate67
}

yystate72:
c = y.getc()
switch {
default:
goto yyabort
case c == '*' || c == '/':
goto yystate66
case c == '\n':
goto yystate56
case c == ']':
goto yystate73
case c >= '\x01' && c <= '\t' || c >= '\v' && c <= ')' || c >= '+' && c <= '.' || c >= '0' && c <= '\\' || c >= '^' && c <= 'ÿ':
goto yystate64
}

yystate73:
c = y.getc()
switch {
default:
goto yyrule5
case c == '/':
goto yystate65
case c == '\n':
goto yystate56
case c == ']':
goto yystate73
case c >= '\x01' && c <= '\t' || c >= '\v' && c <= '.' || c >= '0' && c <= '\\' || c >= '^' && c <= 'ÿ':
goto yystate64
}

yystate74:
c = y.getc()
goto yyrule5

yystate75:
c = y.getc()
switch {
default:
goto yyabort
case c == '/':
goto yystate58
}

yystate76:
c = y.getc()
switch {
default:
goto yyrule4
case c == '=':
goto yystate77
}

yystate77:
c = y.getc()
goto yyrule18

yystate78:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c >= 'c' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 'b':
goto yystate79
case c == 's':
goto yystate86
}

yystate79:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 's':
goto yystate80
}

yystate80:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate81
}

yystate81:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'r':
goto yystate82
}

yystate82:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate83
}

yystate83:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
goto yystate2
case c == 'c':
goto yystate84
}

yystate84:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate85
}

yystate85:
c = y.getc()
switch {
default:
goto yyrule20
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate86:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 's':
goto yystate87
}

yystate87:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate88
}

yystate88:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'r':
goto yystate89
}

yystate89:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate90
}

yystate90:
c = y.getc()
switch {
default:
goto yyrule67
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate91:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c == 'p' || c == 'q' || c >= 's' && c <= 'x' || c == 'z':
goto yystate2
case c == 'o':
goto yystate92
case c == 'r':
goto yystate98
case c == 'y':
goto yystate102
}

yystate92:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'o':
goto yystate93
}

yystate93:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate94
}

yystate94:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate95
}

yystate95:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate96
}

yystate96:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate97
}

yystate97:
c = y.getc()
switch {
default:
goto yyrule25
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate98:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate99
}

yystate99:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate100
}

yystate100:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'j' || c >= 'l' && c <= 'z':
goto yystate2
case c == 'k':
goto yystate101
}

yystate101:
c = y.getc()
switch {
default:
goto yyrule30
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate102:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate103
}

yystate103:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate104
}

yystate104:
c = y.getc()
switch {
default:
goto yyrule34
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate105:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'g' || c >= 'i' && c <= 'k' || c == 'm' || c == 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate106
case c == 'h':
goto yystate112
case c == 'l':
goto yystate115
case c == 'o':
goto yystate119
}

yystate106:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 'u' && c <= 'z':
goto yystate2
case c == 's':
goto yystate107
case c == 't':
goto yystate109
}

yystate107:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate108
}

yystate108:
c = y.getc()
switch {
default:
goto yyrule39
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate109:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
goto yystate2
case c == 'c':
goto yystate110
}

yystate110:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'g' || c >= 'i' && c <= 'z':
goto yystate2
case c == 'h':
goto yystate111
}

yystate111:
c = y.getc()
switch {
default:
goto yyrule46
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate112:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate113
}

yystate113:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'r':
goto yystate114
}

yystate114:
c = y.getc()
switch {
default:
goto yyrule51
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate115:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate116
}

yystate116:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 's':
goto yystate117
}

yystate117:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 's':
goto yystate118
}

yystate118:
c = y.getc()
switch {
default:
goto yyrule56
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate119:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate120
}

yystate120:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate121
}

yystate121:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate122
}

yystate122:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate123
}

yystate123:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 't' || c >= 'v' && c <= 'z':
goto yystate2
case c == 'u':
goto yystate124
}

yystate124:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate125
}

yystate125:
c = y.getc()
switch {
default:
goto yyrule61
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate126:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate127
case c == 'o':
goto yystate133
}

yystate127:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'e' || c >= 'g' && c <= 'z':
goto yystate2
case c == 'f':
goto yystate128
}

yystate128:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate129
}

yystate129:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 't' || c >= 'v' && c <= 'z':
goto yystate2
case c == 'u':
goto yystate130
}

yystate130:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate131
}

yystate131:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate132
}

yystate132:
c = y.getc()
switch {
default:
goto yyrule63
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate133:
c = y.getc()
switch {
default:
goto yyrule21
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 't' || c >= 'v' && c <= 'z':
goto yystate2
case c == 'u':
goto yystate134
}

yystate134:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c >= 'c' && c <= 'z':
goto yystate2
case c == 'b':
goto yystate135
}

yystate135:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate136
}

yystate136:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate137
}

yystate137:
c = y.getc()
switch {
default:
goto yyrule26
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate138:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c == 'm' || c >= 'o' && c <= 'w' || c == 'y' || c == 'z':
goto yystate2
case c == 'l':
goto yystate139
case c == 'n':
goto yystate142
case c == 'x':
goto yystate145
}

yystate139:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 's':
goto yystate140
}

yystate140:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate141
}

yystate141:
c = y.getc()
switch {
default:
goto yyrule31
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate142:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 't' || c >= 'v' && c <= 'z':
goto yystate2
case c == 'u':
goto yystate143
}

yystate143:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'l' || c >= 'n' && c <= 'z':
goto yystate2
case c == 'm':
goto yystate144
}

yystate144:
c = y.getc()
switch {
default:
goto yyrule66
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate145:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate146
}

yystate146:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate147
}

yystate147:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate148
}

yystate148:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'z':
goto yystate2
case c == 'd':
goto yystate149
}

yystate149:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 's':
goto yystate150
}

yystate150:
c = y.getc()
switch {
default:
goto yyrule35
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate151:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'h' || c == 'j' || c == 'k' || c == 'm' || c == 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate152
case c == 'i':
goto yystate156
case c == 'l':
goto yystate162
case c == 'o':
goto yystate166
}

yystate152:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate153
}

yystate153:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 's':
goto yystate154
}

yystate154:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate155
}

yystate155:
c = y.getc()
switch {
default:
goto yyrule2
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate156:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate157
}

yystate157:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate158
}

yystate158:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate159
}

yystate159:
c = y.getc()
switch {
default:
goto yyrule40
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate160
}

yystate160:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'x' || c == 'z':
goto yystate2
case c == 'y':
goto yystate161
}

yystate161:
c = y.getc()
switch {
default:
goto yyrule42
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate162:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'o':
goto yystate163
}

yystate163:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate164
}

yystate164:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate165
}

yystate165:
c = y.getc()
switch {
default:
goto yyrule47
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate166:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'r':
goto yystate167
}

yystate167:
c = y.getc()
switch {
default:
goto yyrule52
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate168:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'e' || c >= 'g' && c <= 'l' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'f':
goto yystate169
case c == 'm':
goto yystate170
case c == 'n':
goto yystate182
}

yystate169:
c = y.getc()
switch {
default:
goto yyrule64
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate170:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'o' || c >= 'q' && c <= 'z':
goto yystate2
case c == 'p':
goto yystate171
}

yystate171:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c == 'm' || c == 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate172
case c == 'o':
goto yystate179
}

yystate172:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate173
}

yystate173:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'l' || c >= 'n' && c <= 'z':
goto yystate2
case c == 'm':
goto yystate174
}

yystate174:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate175
}

yystate175:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate176
}

yystate176:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate177
}

yystate177:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 's':
goto yystate178
}

yystate178:
c = y.getc()
switch {
default:
goto yyrule22
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate179:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'r':
goto yystate180
}

yystate180:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate181
}

yystate181:
c = y.getc()
switch {
default:
goto yyrule27
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate182:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 'u' && c <= 'z':
goto yystate2
case c == 's':
goto yystate183
case c == 't':
goto yystate191
}

yystate183:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate184
}

yystate184:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate185
}

yystate185:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate186
}

yystate186:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
goto yystate2
case c == 'c':
goto yystate187
}

yystate187:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate188
}

yystate188:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'o':
goto yystate189
}

yystate189:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'e' || c >= 'g' && c <= 'z':
goto yystate2
case c == 'f':
goto yystate190
}

yystate190:
c = y.getc()
switch {
default:
goto yyrule36
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate191:
c = y.getc()
switch {
default:
goto yyrule41
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate192
}

yystate192:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'r':
goto yystate193
}

yystate193:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'e' || c >= 'g' && c <= 'z':
goto yystate2
case c == 'f':
goto yystate194
}

yystate194:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate195
}

yystate195:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
goto yystate2
case c == 'c':
goto yystate196
}

yystate196:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate197
}

yystate197:
c = y.getc()
switch {
default:
goto yyrule43
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate198:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'o':
goto yystate199
}

yystate199:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate200
}

yystate200:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'f' || c >= 'h' && c <= 'z':
goto yystate2
case c == 'g':
goto yystate201
}

yystate201:
c = y.getc()
switch {
default:
goto yyrule48
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate202:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'd' || c >= 'f' && c <= 't' || c >= 'v' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate203
case c == 'e':
goto yystate208
case c == 'u':
goto yystate210
}

yystate203:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate204
}

yystate204:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate205
}

yystate205:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'u' || c >= 'w' && c <= 'z':
goto yystate2
case c == 'v':
goto yystate206
}

yystate206:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate207
}

yystate207:
c = y.getc()
switch {
default:
goto yyrule53
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate208:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'v' || c >= 'x' && c <= 'z':
goto yystate2
case c == 'w':
goto yystate209
}

yystate209:
c = y.getc()
switch {
default:
goto yyrule57
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate210:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate211
}

yystate211:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate212
}

yystate212:
c = y.getc()
switch {
default:
goto yyrule59
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate213:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'q' || c == 's' || c == 't' || c >= 'v' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate214
case c == 'r':
goto yystate220
case c == 'u':
goto yystate233
}

yystate214:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
goto yystate2
case c == 'c':
goto yystate215
}

yystate215:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'j' || c >= 'l' && c <= 'z':
goto yystate2
case c == 'k':
goto yystate216
}

yystate216:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate217
}

yystate217:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'f' || c >= 'h' && c <= 'z':
goto yystate2
case c == 'g':
goto yystate218
}

yystate218:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate219
}

yystate219:
c = y.getc()
switch {
default:
goto yyrule23
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate220:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate221
case c == 'o':
goto yystate226
}

yystate221:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'u' || c >= 'w' && c <= 'z':
goto yystate2
case c == 'v':
goto yystate222
}

yystate222:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate223
}

yystate223:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate224
}

yystate224:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate225
}

yystate225:
c = y.getc()
switch {
default:
goto yyrule28
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate226:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate227
}

yystate227:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate228
}

yystate228:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
goto yystate2
case c == 'c':
goto yystate229
}

yystate229:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate230
}

yystate230:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate231
}

yystate231:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'z':
goto yystate2
case c == 'd':
goto yystate232
}

yystate232:
c = y.getc()
switch {
default:
goto yyrule32
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate233:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c >= 'c' && c <= 'z':
goto yystate2
case c == 'b':
goto yystate234
}

yystate234:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate235
}

yystate235:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate236
}

yystate236:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
goto yystate2
case c == 'c':
goto yystate237
}

yystate237:
c = y.getc()
switch {
default:
goto yyrule37
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate238:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate239
}

yystate239:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate240
}

yystate240:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 't' || c >= 'v' && c <= 'z':
goto yystate2
case c == 'u':
goto yystate241
}

yystate241:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'r':
goto yystate242
}

yystate242:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate243
}

yystate243:
c = y.getc()
switch {
default:
goto yyrule44
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate244:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'g' || c >= 'i' && c <= 's' || c == 'v' || c == 'x' || c == 'z':
goto yystate2
case c == 'h':
goto yystate245
case c == 't':
goto yystate249
case c == 'u':
goto yystate254
case c == 'w':
goto yystate258
case c == 'y':
goto yystate263
}

yystate245:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'o':
goto yystate246
}

yystate246:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'r':
goto yystate247
}

yystate247:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate248
}

yystate248:
c = y.getc()
switch {
default:
goto yyrule49
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate249:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate250
}

yystate250:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate251
}

yystate251:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate252
}

yystate252:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
goto yystate2
case c == 'c':
goto yystate253
}

yystate253:
c = y.getc()
switch {
default:
goto yyrule54
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate254:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'o' || c >= 'q' && c <= 'z':
goto yystate2
case c == 'p':
goto yystate255
}

yystate255:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate256
}

yystate256:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'r':
goto yystate257
}

yystate257:
c = y.getc()
switch {
default:
goto yyrule58
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate258:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate259
}

yystate259:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate260
}

yystate260:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
goto yystate2
case c == 'c':
goto yystate261
}

yystate261:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'g' || c >= 'i' && c <= 'z':
goto yystate2
case c == 'h':
goto yystate262
}

yystate262:
c = y.getc()
switch {
default:
goto yyrule60
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate263:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate264
}

yystate264:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
goto yystate2
case c == 'c':
goto yystate265
}

yystate265:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'g' || c >= 'i' && c <= 'z':
goto yystate2
case c == 'h':
goto yystate266
}

yystate266:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'r':
goto yystate267
}

yystate267:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'o':
goto yystate268
}

yystate268:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate269
}

yystate269:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate270
}

yystate270:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'y':
goto yystate2
case c == 'z':
goto yystate271
}

yystate271:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate272
}

yystate272:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'z':
goto yystate2
case c == 'd':
goto yystate273
}

yystate273:
c = y.getc()
switch {
default:
goto yyrule62
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate274:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'g' || c >= 'i' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'h':
goto yystate275
case c == 'r':
goto yystate282
}

yystate275:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'q' || c >= 's' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate276
case c == 'r':
goto yystate278
}

yystate276:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 's':
goto yystate277
}

yystate277:
c = y.getc()
switch {
default:
goto yyrule65
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate278:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'o':
goto yystate279
}

yystate279:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'v' || c >= 'x' && c <= 'z':
goto yystate2
case c == 'w':
goto yystate280
}

yystate280:
c = y.getc()
switch {
default:
goto yyrule24
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 's':
goto yystate281
}

yystate281:
c = y.getc()
switch {
default:
goto yyrule29
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate282:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 't' || c >= 'v' && c <= 'x' || c == 'z':
goto yystate2
case c == 'a':
goto yystate283
case c == 'u':
goto yystate290
case c == 'y':
goto yystate292
}

yystate283:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate284
}

yystate284:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
goto yystate2
case c == 's':
goto yystate285
}

yystate285:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate286
}

yystate286:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate287
}

yystate287:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
goto yystate2
case c == 'n':
goto yystate288
}

yystate288:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate289
}

yystate289:
c = y.getc()
switch {
default:
goto yyrule33
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate290:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate291
}

yystate291:
c = y.getc()
switch {
default:
goto yyrule1
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate292:
c = y.getc()
switch {
default:
goto yyrule38
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate293:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
goto yystate2
case c == 'o':
goto yystate294
}

yystate294:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c == 'j' || c == 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate295
case c == 'l':
goto yystate297
}

yystate295:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'z':
goto yystate2
case c == 'd':
goto yystate296
}

yystate296:
c = y.getc()
switch {
default:
goto yyrule45
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate297:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
goto yystate2
case c == 'a':
goto yystate298
}

yystate298:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
goto yystate2
case c == 't':
goto yystate299
}

yystate299:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate300
}

yystate300:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate301
}

yystate301:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate302
}

yystate302:
c = y.getc()
switch {
default:
goto yyrule50
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate303:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'g' || c >= 'i' && c <= 'z':
goto yystate2
case c == 'h':
goto yystate304
}

yystate304:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
goto yystate2
case c == 'i':
goto yystate305
}

yystate305:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
goto yystate2
case c == 'l':
goto yystate306
}

yystate306:
c = y.getc()
switch {
default:
goto yyrule68
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
goto yystate2
case c == 'e':
goto yystate307
}

yystate307:
c = y.getc()
switch {
default:
goto yyrule55
case c == '\x01' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
goto yystate2
}

yystate308:
c = y.getc()
switch {
default:
goto yyrule4
case c == '=':
goto yystate309
case c == '|':
goto yystate310
}

yystate309:
c = y.getc()
goto yyrule17

yystate310:
c = y.getc()
goto yyrule8

yyrule1: // "true"
{
	{return y.LexString(BOOLLIT, lval)}
goto yystate0
}
yyrule2: // "false"
{
	{return y.LexString(BOOLLIT, lval)}
goto yystate0
}
yyrule3: // {Separator}
{
{return y.LexChar(lval)}
goto yystate0
}
yyrule4: // {Delimiter1}
{
{return y.LexChar(lval)}
goto yystate0
}
yyrule5: // {Dimension}
{
{return y.LexString(OP_DIM, lval)}
goto yystate0
}
yyrule6: // "=="
{
	{return y.LexString(OP_EQ, lval)}
goto yystate0
}
yyrule7: // "!="
{
	{return y.LexString(OP_NE, lval)}
goto yystate0
}
yyrule8: // "||"
{
	{return y.LexString(OP_LOR, lval)}
goto yystate0
}
yyrule9: // "&&"
{
	{return y.LexString(OP_LAND, lval)}
goto yystate0
}
yyrule10: // "++"
{
	{return y.LexString(OP_INC, lval)}
goto yystate0
}
yyrule11: // "--"
{
	{return y.LexString(OP_DEC, lval)}
goto yystate0
}
yyrule12: // "+="
{
	{return y.LexString(ASS_ADD, lval)}
goto yystate0
}
yyrule13: // "-="
{
	{return y.LexString(ASS_SUB, lval)}
goto yystate0
}
yyrule14: // "*="
{
	{return y.LexString(ASS_MUL, lval)}
goto yystate0
}
yyrule15: // "/="
{
	{return y.LexString(ASS_DIV, lval)}
goto yystate0
}
yyrule16: // "&="
{
	{return y.LexString(ASS_AND, lval)}
goto yystate0
}
yyrule17: // "|="
{
	{return y.LexString(ASS_OR, lval)}
goto yystate0
}
yyrule18: // "^="
{
	{return y.LexString(ASS_XOR, lval)}
goto yystate0
}
yyrule19: // "%="
{
	{return y.LexString(ASS_MOD, lval)}
goto yystate0
}
yyrule20: // "abstract"
{
{return y.LexString(ABSTRACT, lval)}
goto yystate0
}
yyrule21: // "do"
{
           {return y.LexString(DO, lval)}
goto yystate0
}
yyrule22: // "implements"
{
   {return y.LexString(IMPLEMENTS, lval)}
goto yystate0
}
yyrule23: // "package"
{
{return y.LexString(PACKAGE, lval)}
goto yystate0
}
yyrule24: // "throw"
{
	{return y.LexString(THROW, lval)}
goto yystate0
}
yyrule25: // "boolean"
{
{return y.LexString(BOOLEAN, lval)}
goto yystate0
}
yyrule26: // "double"
{
{return y.LexString(DOUBLE, lval)}
goto yystate0
}
yyrule27: // "import"
{
{return y.LexString(IMPORT, lval)}
goto yystate0
}
yyrule28: // "private"
{
{return y.LexString(PRIVATE, lval)}
goto yystate0
}
yyrule29: // "throws"
{
{return y.LexString(THROWS, lval)}
goto yystate0
}
yyrule30: // "break"
{
	{return y.LexString(BREAK, lval)}
goto yystate0
}
yyrule31: // "else"
{
	{return y.LexString(ELSE, lval)}
goto yystate0
}
yyrule32: // "protected"
{
{return y.LexString(PROTECTED, lval)}
goto yystate0
}
yyrule33: // "transient"
{
{return y.LexString(TRANSIENT, lval)}
goto yystate0
}
yyrule34: // "byte"
{
	{return y.LexString(BYTE, lval)}
goto yystate0
}
yyrule35: // "extends"
{
{return y.LexString(EXTENDS, lval)}
goto yystate0
}
yyrule36: // "instanceof"
{
{return y.LexString(INSTANCEOF, lval)}
goto yystate0
}
yyrule37: // "public"
{
{return y.LexString(PUBLIC, lval)}
goto yystate0
}
yyrule38: // "try"
{
	{return y.LexString(TRY, lval)}
goto yystate0
}
yyrule39: // "case"
{
	{return y.LexString(CASE, lval)}
goto yystate0
}
yyrule40: // "final"
{
	{return y.LexString(FINAL, lval)}
goto yystate0
}
yyrule41: // "int"
{
	{return y.LexString(INT, lval)}
goto yystate0
}
yyrule42: // "finally"
{
{return y.LexString(FINALLY, lval)}
goto yystate0
}
yyrule43: // "interface"
{
{return y.LexString(INTERFACE, lval)}
goto yystate0
}
yyrule44: // "return"
{
{return y.LexString(RETURN, lval)}
goto yystate0
}
yyrule45: // "void"
{
	{return y.LexString(VOID, lval)}
goto yystate0
}
yyrule46: // "catch"
{
	{return y.LexString(CATCH, lval)}
goto yystate0
}
yyrule47: // "float"
{
	{return y.LexString(FLOAT, lval)}
goto yystate0
}
yyrule48: // "long"
{
	{return y.LexString(LONG, lval)}
goto yystate0
}
yyrule49: // "short"
{
	{return y.LexString(SHORT, lval)}
goto yystate0
}
yyrule50: // "volatile"
{
{return y.LexString(VOLATILE, lval)}
goto yystate0
}
yyrule51: // "char"
{
	{return y.LexString(CHAR, lval)}
goto yystate0
}
yyrule52: // "for"
{
	{return y.LexString(FOR, lval)}
goto yystate0
}
yyrule53: // "native"
{
{return y.LexString(NATIVE, lval)}
goto yystate0
}
yyrule54: // "static"
{
{return y.LexString(STATIC, lval)}
goto yystate0
}
yyrule55: // "while"
{
	{return y.LexString(WHILE, lval)}
goto yystate0
}
yyrule56: // "class"
{
	{return y.LexString(CLASS, lval)}
goto yystate0
}
yyrule57: // "new"
{
	{return y.LexString(NEW, lval)}
goto yystate0
}
yyrule58: // "super"
{
	{return y.LexString(SUPER, lval)}
goto yystate0
}
yyrule59: // "null"
{
	{return y.LexString(JNULL, lval)}
goto yystate0
}
yyrule60: // "switch"
{
{return y.LexString(SWITCH, lval)}
goto yystate0
}
yyrule61: // "continue"
{
{return y.LexString(CONTINUE, lval)}
goto yystate0
}
yyrule62: // "synchronized"
{
{return y.LexString(SYNCHRONIZED, lval)}
goto yystate0
}
yyrule63: // "default"
{
{return y.LexString(DEFAULT, lval)}
goto yystate0
}
yyrule64: // "if"
{
	{return y.LexString(IF, lval)}
goto yystate0
}
yyrule65: // "this"
{
	{return y.LexString(THIS, lval)}
goto yystate0
}
yyrule66: // "enum"
{
	{return y.LexString(ENUM, lval)}
goto yystate0
}
yyrule67: // "assert"
{
{return y.LexString(ASSERT, lval)}
goto yystate0
}
yyrule68: // {Identifier}
{
{return y.LexString(IDENTIFIER, lval)}
goto yystate0
}
yyrule69: // {Literal}
{
      {return y.LexString(LITERAL, lval)}
goto yystate0
}
yyrule70: // {CR}
{
  		{}
goto yystate0
}
yyrule71: // {LF}
{
	{}
goto yystate0
}
yyrule72: // {FF}
{
	{}
goto yystate0
}
yyrule73: // {TAB}
{
	{}
goto yystate0
}
yyrule74: // {BLK}
{
          {}
goto yystate0
}
yyrule75: // {BLANK}
{
	{}
goto yystate0
}
yyrule76: // {Comment}
{
{}
goto yystate0
}
panic("unreachable")

goto yyabort // silence unused label error

yyabort: // no lexem recognized
	y.empty = true
	return 0
}
