package fins

import (
	"fmt"
)

func (sys *FinsSysTp) FinslibMemoryAreaReadWord(start string, data []byte, num_words uint32) int {

	if num_words == 0 {
		return FINS_RETVAL_SUCCESS
	}
	if sys == nil {
		return FINS_RETVAL_NOT_INITIALIZED
	}

	if start == "" {
		return FINS_RETVAL_NO_READ_ADDRESS
	}
	if data == nil {
		return FINS_RETVAL_NO_DATA_BLOCK
	}

	var address fins_address_tp
	address.name = make([]byte, 4)
	var chunk_start uint32 = address.main_address
	var chunk_length uint32
	var offset uint32 = 0
	var todo uint32 = num_words
	var a uint32
	var bodylen uint32
	var retval uint32
	var fins_cmnd fins_command_tp

	Err := XX_finslib_decode_address(start, &address)
	if Err != nil {
		fmt.Println(Err)
		return FINS_RETVAL_INVALID_READ_ADDRESS
	}

	area_ptr := XX_finslib_search_area(sys, &address, 16, FI_RD, false)
	if area_ptr == nil {
		return FINS_RETVAL_INVALID_READ_AREA
	}

	chunk_start += area_ptr.low_addr >> 8
	chunk_start -= area_ptr.low_id
	for {
		chunk_length = FINS_MAX_READ_WORDS_SYSWAY
		if chunk_length > todo {
			chunk_length = todo
		}

		XX_finslib_init_command(sys, &fins_cmnd, 0x01, 0x01)
		bodylen = 0
		bodylen++
		fins_cmnd.body[bodylen] = uint32(area_ptr.area)
		bodylen++
		fins_cmnd.body[bodylen] = (chunk_start >> 8) & 0xff
		bodylen++
		fins_cmnd.body[bodylen] = (chunk_start) & 0xff
		bodylen++
		fins_cmnd.body[bodylen] = 0x00
		bodylen++
		fins_cmnd.body[bodylen] = (chunk_length >> 8) & 0xff
		bodylen++
		fins_cmnd.body[bodylen] = (chunk_length) & 0xff
		retval = uint32(XX_finslib_communicate(sys, &fins_cmnd, &bodylen))
		if retval != FINS_RETVAL_SUCCESS {
			return int(retval)
		}

		if bodylen != 2+2*chunk_length {
			return FINS_RETVAL_BODY_TOO_SHORT
		}
		bodylen = 2

		for a = 0; a < 2*chunk_length; a++ {
			bodylen++
			data[offset+a] = byte(fins_cmnd.body[bodylen])
		}

		todo -= chunk_length
		offset += chunk_length * 2
		chunk_start += chunk_length
		if todo <= 0 {
			break
		}
	}
	/*

		do {
			chunk_length = FINS_MAX_READ_WORDS_SYSWAY;
			if ( chunk_length > todo ) chunk_length = todo;

			XX_finslib_init_command( sys, & fins_cmnd, 0x01, 0x01 );

			bodylen = 0;

			fins_cmnd.body[bodylen++] = area_ptr->area;
			fins_cmnd.body[bodylen++] = (chunk_start  >> 8) & 0xff;
			fins_cmnd.body[bodylen++] = (chunk_start      ) & 0xff;
			fins_cmnd.body[bodylen++] = 0x00;
			fins_cmnd.body[bodylen++] = (chunk_length >> 8) & 0xff;
			fins_cmnd.body[bodylen++] = (chunk_length     ) & 0xff;

			if ( ( retval = XX_finslib_communicate( sys, & fins_cmnd, & bodylen ) ) != FINS_RETVAL_SUCCESS ) return retval;

			if ( bodylen != 2+2*chunk_length ) return FINS_RETVAL_BODY_TOO_SHORT;

			bodylen = 2;

			for (a=0; a<2*chunk_length; a++) data[offset+a] = fins_cmnd.body[bodylen++];

			todo        -= chunk_length;
			offset      += chunk_length * 2;
			chunk_start += chunk_length;

		} while ( todo > 0 );
	*/
	return FINS_RETVAL_SUCCESS
}
