package dns

import (
	"errors"
	"fmt"
)

var (
	ErrBuf        error = errors.New("dns: buffer size too small")
	ErrRdata      error = errors.New("dns: bad rdata")
	ErrLongDomain error = errors.New(fmt.Sprintf("domain name exceeded %d wire-format octets", maxDomainNameWireOctets))
	ErrCompress   error = errors.New("too many compression pointers")
)

const (
	escapedByteSmall = "" +
		`\000\001\002\003\004\005\006\007\008\009` +
		`\010\011\012\013\014\015\016\017\018\019` +
		`\020\021\022\023\024\025\026\027\028\029` +
		`\030\031`
	escapedByteLarge = `\127\128\129` +
		`\130\131\132\133\134\135\136\137\138\139` +
		`\140\141\142\143\144\145\146\147\148\149` +
		`\150\151\152\153\154\155\156\157\158\159` +
		`\160\161\162\163\164\165\166\167\168\169` +
		`\170\171\172\173\174\175\176\177\178\179` +
		`\180\181\182\183\184\185\186\187\188\189` +
		`\190\191\192\193\194\195\196\197\198\199` +
		`\200\201\202\203\204\205\206\207\208\209` +
		`\210\211\212\213\214\215\216\217\218\219` +
		`\220\221\222\223\224\225\226\227\228\229` +
		`\230\231\232\233\234\235\236\237\238\239` +
		`\240\241\242\243\244\245\246\247\248\249` +
		`\250\251\252\253\254\255`
)

// escapeByte returns the \DDD escaping of b which must
// satisfy b < ' ' || b > '~'.
func escapeByte(b byte) string {
	if b < ' ' {
		return escapedByteSmall[b*4 : b*4+4]
	}

	b -= '~' + 1
	// The cast here is needed as b*4 may overflow byte.
	return escapedByteLarge[int(b)*4 : int(b)*4+4]
}

// isDomainNameLabelSpecial returns true if
// a domain name label byte should be prefixed
// with an escaping backslash.
func isDomainNameLabelSpecial(b byte) bool {
	switch b {
	case '.', ' ', '\'', '@', ';', '(', ')', '"', '\\':
		return true
	}
	return false
}

func unpackDomainName(msg []byte, off int) (string, int, error) {
	s := make([]byte, 0, maxDomainNamePresentationLength)
	off1 := 0
	lenmsg := len(msg)
	budget := maxDomainNameWireOctets
	ptr := 0 // number of pointers followed
Loop:
	for {
		if off >= lenmsg {
			return "", lenmsg, ErrBuf
		}
		c := int(msg[off])
		off++
		switch c & 0xC0 {
		case 0x00:
			if c == 0x00 {
				// end of name
				break Loop
			}
			// literal string
			if off+c > lenmsg {
				return "", lenmsg, ErrBuf
			}
			budget -= c + 1 // +1 for the label separator
			if budget <= 0 {
				return "", lenmsg, ErrLongDomain
			}
			for _, b := range msg[off : off+c] {
				if isDomainNameLabelSpecial(b) {
					s = append(s, '\\', b)
				} else if b < ' ' || b > '~' {
					s = append(s, escapeByte(b)...)
				} else {
					s = append(s, b)
				}
			}
			s = append(s, '.')
			off += c
		case 0xC0:
			// pointer to somewhere else in msg.
			// remember location after first ptr,
			// since that's how many bytes we consumed.
			// also, don't follow too many pointers --
			// maybe there's a loop.
			if off >= lenmsg {
				return "", lenmsg, ErrBuf
			}
			c1 := msg[off]
			off++
			if ptr == 0 {
				off1 = off
			}
			if ptr++; ptr > maxCompressionPointers {
				return "", lenmsg, ErrCompress
			}
			// pointer should guarantee that it advances and points forwards at least
			// but the condition on previous three lines guarantees that it's
			// at least loop-free
			off = (c^0xC0)<<8 | int(c1)
		default:
			// 0x80 and 0x40 are reserved
			return "", lenmsg, ErrRdata
		}
	}
	if ptr == 0 {
		off1 = off
	}
	if len(s) == 0 {
		return ".", off1, nil
	}
	return string(s), off1, nil
}
