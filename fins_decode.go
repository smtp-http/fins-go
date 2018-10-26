package fins

import (
	"errors"
	"fmt"
	"strings"
)

func XX_finslib_decode_address(str string, address *fins_address_tp) error {
	if str == "" || address == nil {
		return errors.New("error: start is NULL or address is NULL")
	}
	len := len(str)
	ptr := make([]byte, 100)
	name := make([]byte, 10)
	//nc :=
	copy(ptr, str) //[]byte(str)

	num_char := 0
	var main_address uint32 = 0
	var sub_address uint32 = 0
	i := 0
	for {
		if isSpace(ptr[i]) == false {
			break
		}
		i++
	}
	for {
		if isalpha(ptr[i]) == false || num_char >= 3 {
			break
		}
		name[num_char] = Toupper(ptr[i])
		num_char++
		i++
	}
	if isalpha(ptr[i]) == true {
		return errors.New("error: the layout of start is false, there is more than 3 char")
	}
	for {
		if num_char >= 4 {
			break
		}
		num_char++
		name[num_char] = 0
	}
	for {
		if isSpace(ptr[i]) == false {
			break
		}
		i++
	}
	if i > len {
		return errors.New("error: the layout of start is false, out of range")
	}
	if isdigit(ptr[i]) == false {
		return errors.New("error: the layout of start is false, there is no number")
	}
	for {
		if isdigit(ptr[i]) == false {
			break
		}
		main_address = 10 * main_address
		main_address += uint32(ptr[i]) - '0'
		i++
	}
	if ptr[i] == '.' {

		i++
		for {
			if isdigit(ptr[i]) == false {
				break
			}
			sub_address = 10 * sub_address
			sub_address += uint32(ptr[i]) - '0'
			i++
		}
		if sub_address > 15 {
			return errors.New("error: the layout of start is false, Postdecimal value is greater than 15")
		}
	}
	for {
		if isdigit(ptr[i]) == false {
			break
		}
		i++
	}

	if ptr[i] != 0 {
		return errors.New("error: the layout of start is false, the last one is not The blank space")
	}
	fmt.Printf("name[0]:%v  name[1]:%v  name[2]:%v  name[3]:%v  name[4]:%v  \n", name[0], name[1], name[2], name[3], name[4])
	address.name[0] = name[0]
	address.name[1] = name[1]
	address.name[2] = name[2]
	address.name[3] = name[3]

	address.main_address = main_address
	address.sub_address = sub_address

	return nil
}

/*
	size_t num_char;
	uint32_t main_address;
	uint32_t sub_address;
	char name[4];
	const char *ptr;

	if ( str == NULL  ||  address == NULL ) return true;

	num_char     = 0;
	ptr          = str;
	main_address = 0;
	sub_address  = 0;

	while ( isspace( *ptr ) ) ptr++;

	while ( isalpha( *ptr )  &&  num_char < 3 ) {

		name[num_char] = (char) toupper( *ptr );
		num_char++;
		ptr++;
	}
	if ( isalpha( *ptr ) ) return true;

	while ( num_char < 4 ) name[num_char++] = 0;

	while ( isspace( *ptr ) ) ptr++;
	if ( ! isdigit( *ptr ) ) return true;

	while ( isdigit( *ptr ) ) {

		main_address *= 10;
		main_address += *ptr-'0';
		ptr++;
	}

	if ( *ptr == '.' ) {

		ptr++;
		while ( isdigit( *ptr ) ) {

			sub_address *= 10;
			sub_address += *ptr-'0';
			ptr++;
		}

		if ( sub_address > 15 ) return true;
	}

	while ( isdigit( *ptr ) ) ptr++;

	if ( *ptr ) return true;

	address->name[0]      = name[0];
	address->name[1]      = name[1];
	address->name[2]      = name[2];
	address->name[3]      = name[3];

	address->main_address = main_address;
	address->sub_address  = sub_address;

	return false;
*/

func isSpace(num byte) bool {
	if num == '\t' || num == '\n' || num == ' ' {
		return true
	} else {
		return false
	}
}

func isalpha(num byte) bool {
	if num >= 'a' && num <= 'z' || num >= 'A' && num <= 'Z' {
		return true
	} else {
		return false
	}
}

func isdigit(num byte) bool {
	if num >= '0' && num <= '9' {
		return true
	} else {
		return false
	}
}

func Toupper(num byte) byte {
	if num >= 'a' && num <= 'z' {
		return num - ('a' - 'A')
	} else {
		return num
	}
}

func XX_finslib_search_area(sys *FinsSysTp, address *fins_address_tp, bits int32, accs int32, force bool) *fins_area_tp {
	var a int = 0

	fmt.Printf("plc_mode:%v   bits:%v   accs:%v    force:%v  main_add:%v   name:%v\n", sys.PlcMode, bits, accs, force, address.main_address, address.name)
	for {
		if fins_area[a].plc_mode == FINS_MODE_UNKNOWN {
			break
		}

		if fins_area[a].plc_mode != sys.PlcMode {
			a++
			continue
		}
		if fins_area[a].bits != bits {
			a++
			continue
		}
		if fins_area[a].access|accs == 0x00000000 {
			a++
			continue
		}
		if fins_area[a].force != force {
			a++
			continue
		}
		if fins_area[a].low_id > address.main_address {
			a++
			continue
		}
		if fins_area[a].high_id < address.main_address {
			a++
			continue
		}
		if strings.Compare(fins_area[a].name, string(address.name)) == 1 {
			a++
			continue
		}
		break
	}

	fmt.Println("fins_area[a]: ", &fins_area[a])

	if fins_area[a].plc_mode == FINS_MODE_UNKNOWN {
		return nil
	}

	return &fins_area[a] //&fins_area_tp[a]

}

var fins_area = []fins_area_tp{
	/* plc_mode     name,  bits, length, area, low_id, high_id, low_addr, high_addr, access,                                                      force */
	{FINS_MODE_CS, "CIO", 1, 1, 0x30, 0, 6143, 0x000000, 0x17FF0F, FI_RD | FI_WR | FI_MRD | FI_FRC, false},
	{FINS_MODE_CS, "CIO", 1, 1, 0x70, 0, 6143, 0x000000, 0x17FF0F, FI_MRD, true},
	{FINS_MODE_CS, "CIO", 16, 2, 0xB0, 0, 6143, 0x000000, 0x17FF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "CIO", 16, 4, 0xF0, 0, 6143, 0x000000, 0x17FF00, FI_MRD, true},
	{FINS_MODE_CS, "W", 1, 1, 0x31, 0, 511, 0x000000, 0x01FF0F, FI_RD | FI_WR | FI_MRD | FI_FRC, false},
	{FINS_MODE_CS, "W", 1, 1, 0x71, 0, 511, 0x000000, 0x01FF0F, FI_MRD, true},
	{FINS_MODE_CS, "W", 16, 2, 0xB1, 0, 511, 0x000000, 0x01FF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "W", 16, 4, 0xF1, 0, 511, 0x000000, 0x01FF00, FI_MRD, true},
	{FINS_MODE_CS, "H", 1, 1, 0x32, 0, 511, 0x000000, 0x01FF0F, FI_RD | FI_WR | FI_MRD | FI_FRC, false},
	{FINS_MODE_CS, "H", 1, 1, 0x72, 0, 511, 0x000000, 0x01FF0F, FI_MRD, true},
	{FINS_MODE_CS, "H", 16, 2, 0xB2, 0, 511, 0x000000, 0x01FF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "H", 16, 4, 0xF2, 0, 511, 0x000000, 0x01FF00, FI_MRD, true},
	{FINS_MODE_CS, "A", 1, 1, 0x33, 0, 959, 0x000000, 0x03BF0F, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "A", 1, 1, 0x33, 448, 959, 0x01C000, 0x03BF0F, FI_WR, false},
	{FINS_MODE_CS, "A", 16, 2, 0xB3, 0, 959, 0x000000, 0x03BF00, FI_RD | FI_MRD | FI_TRS, false},
	{FINS_MODE_CS, "A", 16, 2, 0xB3, 448, 959, 0x01C000, 0x03BF00, FI_WR | FI_FILL | FI_TRD, false},
	{FINS_MODE_CS, "TIM", 1, 1, 0x09, 0, 4095, 0x000000, 0x0FFF00, FI_RD | FI_MRD | FI_FRC, false},
	{FINS_MODE_CS, "TIM", 1, 1, 0x49, 0, 4095, 0x000000, 0x0FFF00, FI_MRD, true},
	{FINS_MODE_CS, "TIM", 16, 2, 0x89, 0, 4095, 0x000000, 0x0FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "CNT", 1, 1, 0x09, 0, 4095, 0x800000, 0x8FFF00, FI_RD | FI_MRD | FI_FRC, false},
	{FINS_MODE_CS, "CNT", 1, 1, 0x49, 0, 4095, 0x800000, 0x8FFF00, FI_MRD, true},
	{FINS_MODE_CS, "CNT", 16, 2, 0x89, 0, 4095, 0x800000, 0x8FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "DM", 1, 1, 0x02, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "DM", 16, 2, 0x82, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "E0_", 1, 1, 0x20, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "E1_", 1, 1, 0x21, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "E2_", 1, 1, 0x22, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "E3_", 1, 1, 0x23, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "E4_", 1, 1, 0x24, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "E5_", 1, 1, 0x25, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "E6_", 1, 1, 0x26, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "E7_", 1, 1, 0x27, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "E8_", 1, 1, 0x28, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "E9_", 1, 1, 0x29, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "EA_", 1, 1, 0x2A, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "EB_", 1, 1, 0x2B, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "EC_", 1, 1, 0x2C, 0, 32767, 0x000000, 0x7FFF0F, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "E0_", 16, 2, 0xA0, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "E1_", 16, 2, 0xA1, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "E2_", 16, 2, 0xA2, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "E3_", 16, 2, 0xA3, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "E4_", 16, 2, 0xA4, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "E5_", 16, 2, 0xA5, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "E6_", 16, 2, 0xA6, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "E7_", 16, 2, 0xA7, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "E8_", 16, 2, 0xA8, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "E9_", 16, 2, 0xA9, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "EA_", 16, 2, 0xAA, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "EB_", 16, 2, 0xAB, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "EC_", 16, 2, 0xAC, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "E", 16, 2, 0x98, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CS, "EM", 16, 2, 0xBC, 0, 0, 0x0F0000, 0x0F0000, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "TK", 1, 1, 0x06, 0, 31, 0x000000, 0x001F00, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "TKS", 1, 1, 0x46, 0, 31, 0x000000, 0x001F00, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "IR", 16, 4, 0xDC, 0, 15, 0x010000, 0x010F00, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "DR", 16, 2, 0xBC, 0, 15, 0x020000, 0x020F00, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_CS, "C1M", 1, 1, 0x07, 0, 0, 0x000000, 0x000000, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "C1S", 1, 1, 0x07, 0, 0, 0x000100, 0x000100, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "C02", 1, 1, 0x07, 0, 0, 0x000200, 0x000200, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "C01", 1, 1, 0x07, 0, 0, 0x000300, 0x000300, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "C22", 1, 1, 0x07, 0, 0, 0x000400, 0x000400, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "CER", 1, 1, 0x07, 0, 0, 0x100000, 0x100000, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "CCY", 1, 1, 0x07, 0, 0, 0x100100, 0x100100, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "CGT", 1, 1, 0x07, 0, 0, 0x100200, 0x100200, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "CEQ", 1, 1, 0x07, 0, 0, 0x100300, 0x100300, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "CLT", 1, 1, 0x07, 0, 0, 0x100400, 0x100400, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "NEG", 1, 1, 0x07, 0, 0, 0x100500, 0x100500, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "COF", 1, 1, 0x07, 0, 0, 0x100600, 0x100600, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "CUF", 1, 1, 0x07, 0, 0, 0x100700, 0x100700, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "CGE", 1, 1, 0x07, 0, 0, 0x100800, 0x100800, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "CNE", 1, 1, 0x07, 0, 0, 0x100900, 0x100900, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "CLE", 1, 1, 0x07, 0, 0, 0x100A00, 0x100A00, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "OFF", 1, 1, 0x07, 0, 0, 0x100E00, 0x100E00, FI_RD | FI_MRD, false},
	{FINS_MODE_CS, "ON", 1, 1, 0x07, 0, 0, 0x100F00, 0x100F00, FI_RD | FI_MRD, false},
	{FINS_MODE_CV, "CIO", 1, 1, 0x00, 0, 2555, 0x000000, 0x09FB0F, FI_RD | FI_MRD | FI_FRC, false},
	{FINS_MODE_CV, "CIO", 1, 1, 0x40, 0, 2555, 0x000000, 0x09FB0F, FI_MRD, true},
	{FINS_MODE_CV, "CIO", 16, 2, 0x80, 0, 2555, 0x000000, 0x09FB00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "CIO", 16, 2, 0xC0, 0, 2555, 0x000000, 0x09FB00, FI_MRD, true},
	{FINS_MODE_CV, "A", 1, 1, 0x00, 0, 959, 0x0B0000, 0x0EBF0F, FI_RD | FI_MRD, false},
	{FINS_MODE_CV, "A", 1, 1, 0x00, 448, 959, 0x0CC000, 0x0EBF0F, 0, false},
	{FINS_MODE_CV, "A", 16, 2, 0x80, 0, 959, 0x0B0000, 0x0EBF00, FI_RD | FI_MRD | FI_TRS, false},
	{FINS_MODE_CV, "A", 16, 2, 0x80, 448, 959, 0x0CC000, 0x0EBF00, FI_WR | FI_FILL | FI_TRD, false},
	{FINS_MODE_CV, "TIM", 1, 1, 0x01, 0, 2047, 0x000000, 0x07FF00, FI_RD | FI_MRD | FI_FRC, false},
	{FINS_MODE_CV, "TIM", 1, 1, 0x41, 0, 2047, 0x000000, 0x07FF00, FI_MRD, true},
	{FINS_MODE_CV, "TIM", 16, 2, 0x81, 0, 2047, 0x000000, 0x07FF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "CNT", 1, 1, 0x01, 0, 2047, 0x080000, 0x0FFF00, FI_RD | FI_MRD | FI_FRC, false},
	{FINS_MODE_CV, "CNT", 1, 1, 0x41, 0, 2047, 0x080000, 0x0FFF00, FI_MRD, true},
	{FINS_MODE_CV, "CNT", 16, 2, 0x81, 0, 2047, 0x080000, 0x0FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "DM", 16, 2, 0x82, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "E0_", 16, 2, 0x90, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "E1_", 16, 2, 0x91, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "E2_", 16, 2, 0x92, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "E3_", 16, 2, 0x93, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "E4_", 16, 2, 0x94, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "E5_", 16, 2, 0x95, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "E6_", 16, 2, 0x96, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "E7_", 16, 2, 0x97, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "E", 16, 2, 0x98, 0, 32767, 0x000000, 0x7FFF00, FI_RD | FI_WR | FI_FILL | FI_MRD | FI_TRS | FI_TRD, false},
	{FINS_MODE_CV, "EM", 16, 2, 0x9C, 0, 0, 0x000600, 0x000600, FI_RD | FI_MRD, false},
	{FINS_MODE_CV, "DR", 16, 2, 0x9C, 0, 2, 0x000300, 0x000500, FI_RD | FI_WR | FI_MRD, false},
	{FINS_MODE_UNKNOWN, "", 0, 0, 0x00, 0, 0, 0x000000, 0x000000, 0, false},
}

func XX_finslib_init_command(sys *FinsSysTp, command *fins_command_tp, mrc uint8, src uint8) {

	command.header[FINS_ICF] = 0x80
	command.header[FINS_RSV] = 0x00
	command.header[FINS_GCT] = 0x02
	command.header[FINS_DNA] = sys.RemoteNet
	command.header[FINS_DA1] = sys.RemoteNode
	command.header[FINS_DA2] = sys.RemoteUnit
	command.header[FINS_SNA] = sys.LocalNet
	command.header[FINS_SA1] = sys.LocalNode
	command.header[FINS_SA2] = sys.LocalUnit
	sys.Sid++
	command.header[FINS_SID] = sys.Sid
	command.header[FINS_MRC] = mrc
	command.header[FINS_SRC] = src
	return
}
