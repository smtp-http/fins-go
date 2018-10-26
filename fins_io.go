package fins

import (
	//	"bytes"
	//	"encoding/binary"
	"errors"
	"fmt"

	//	"net"
	//	"sync"
	//	"sync/atomic"
	"time"
)

////////////////////////////////////////////////////// connect ////////////////////////////////////////////////
func init_system(sys *FinsSysTp) {
	//timeout_val = finslib_monotonic_sec_timer() - 2*FINS_TIMEOUT;
	time_val := time.Now().Unix() - 2*FINS_TIMEOUT
	//if ( finslib_monotonic_sec_timer() > timeout_val ) timeout_val = 0;
	if time.Now().Unix() > time_val {
		time_val = 0
	}

	sys.Address = append(sys.Address, 0)
	//	fmt.Printf("len=%d cap=%d slice=%v\n", len(sys.Address), cap(sys.Address), sys.Address)
	sys.Port = FINS_DEFAULT_PORT
	sys.SocketFd = INVALID_SOCKET
	//sys.Timeout       = timeout_val;
	sys.PlcMode = FINS_MODE_UNKNOWN
	sys.Model = append(sys.Model, 0)
	sys.Version = append(sys.Version, 0)
	sys.Sid = 0
	sys.CommType = FINS_COMM_TYPE_UNKNOWN
	sys.LocalNet = 0
	sys.LocalNode = 0
	sys.LocalUnit = 0
	sys.RemoteNet = 0
	sys.RemoteNode = 0
	sys.RemoteUnit = 0

	sys.Session = nil

	sys.Timeout = time_val
	sys.CliGroup = GetClientGroup()

} /* init_system */

func (s *FinsSysTp) FinslibTcpConnect(address string, port uint16, local_net uint8, local_node uint8, local_unit uint8, remote_net uint8, remote_node uint8, remote_unit uint8, error_val *int32, error_max int32) (*ClientInfo, error) {
	//*error_val = 12
	if time.Now().Unix() < s.Timeout+FINS_TIMEOUT && s.Timeout > 0 {

		if error_val != nil {
			*error_val = FINS_RETVAL_TRY_LATER
		}

		fmt.Println("===== FINS_RETVAL_TRY_LATER! ========")

		return nil, errors.New("FINS_RETVAL_TRY_LATER")
	}

	if port < FINS_PORT_RESERVED || port >= FINS_PORT_MAX {
		port = FINS_DEFAULT_PORT
	}

	addr := []byte(address)
	fmt.Printf("len=%d cap=%d slice=%v\n", len(addr), cap(addr), addr)

	if address == "" || addr[0] == 0 {
		if error_val != nil {
			*error_val = FINS_RETVAL_NO_READ_ADDRESS
		}
		fmt.Println("===== FINS_RETVAL_NO_READ_ADDRESS! ========")
		return nil, errors.New("FINS_RETVAL_NO_READ_ADDRESS")
	}

	init_system(s)

	s.CommType = FINS_COMM_TYPE_TCP
	s.Port = port
	s.LocalNet = local_net
	s.LocalNode = local_node
	s.LocalUnit = local_unit
	s.RemoteNet = remote_net
	s.RemoteNode = remote_node
	s.RemoteUnit = remote_unit

	s.Address = make([]byte, len(addr))
	copy(s.Address, addr)

	strPort := fmt.Sprintf("%d", port)
	addrInfo := address + ":" + strPort
	fmt.Println("addrinfo: ", addrInfo)

	cliAddr := &ClientAddr{}
	//go s.Dial("tcp", addrInfo)
	cliAddr.Ip = string(addr)
	cliAddr.Port = uint(port)
	cliGroup := GetClientGroup()
	err, cliInfo := cliGroup.AddNewClient(cliAddr, error_max)
	if err != nil {
		return nil, errors.New("Connect error!")
	}

	s.Session = cliInfo.Session

	frame := make([]byte, 20)

	frame[0] = 'F' /* Header				*/
	frame[1] = 'I' /*					*/
	frame[2] = 'N' /*					*/
	frame[3] = 'S' /*					*/
	/*					*/
	frame[4] = 0x00  /* Length				*/
	frame[5] = 0x00  /*					*/
	frame[6] = 0x00  /*					*/
	frame[7] = 8 + 4 /*					*/
	/*					*/
	frame[8] = 0x00  /* Command				*/
	frame[9] = 0x00  /*					*/
	frame[10] = 0x00 /*					*/
	frame[11] = 0x00 /*					*/
	/*					*/
	frame[12] = 0x00 /* Error Code	*/
	frame[13] = 0x00 /*					*/
	frame[14] = 0x00 /*					*/
	frame[15] = 0x00 /*					*/
	/*					*/
	frame[16] = 0x00        /* Client node add			*/
	frame[17] = 0x00        /*					*/
	frame[18] = 0x00        /*					*/
	frame[19] = s.LocalNode /* Get node number automatically	*/

	session := cliInfo.Session
	fmt.Printf("session: %v\n", cliInfo.Session)
	err = session.Write(frame)
	if err != nil {
		fmt.Printf("tcp sent error: %v\n", err)
		return nil, err
	}

	n, buf, err := session.ReadData()
	if err != nil {
		fmt.Println("err: ", err)
		return nil, err
	}

	fmt.Println("recv: ", buf, "   n: ", n)
	/*
		n, err := session.conn.Read(buffer)
			ioBuffer.PutBytes(buffer[:n])
			if err != nil {
				session.serv.filterChain.errorCaught(session, err)
				session.Close()
				return nil,err
			}*/
	var command uint32
	command = uint32(buf[8])*256*256*256 + uint32(buf[9])*256*256 + uint32(buf[10])*256 + uint32(buf[11])
	fmt.Println("command: ", command)

	var errorcode uint32
	errorcode = uint32(buf[12])*256*256*256 + uint32(buf[13])*256*256 + uint32(buf[14])*256 + uint32(buf[15])
	fmt.Println("errorcode: ", errorcode)

	/*
		if ( command != 0x00000001 ) {

			new_error          = tcp_errorcode_to_fins_retval( errorcode );
			sys->error_changed = ( new_error != sys->last_error );
			sys->last_error    = new_error;

			if ( error_val != NULL ) *error_val = sys->last_error;

			return fins_close_socket( sys );
		}

	*/

	if command != 0x00000001 {

		new_error := TcpErrorcodeToFinsRetval(errorcode)
		s.errorchange = (new_error != s.lasterror)
		s.lasterror = new_error

		if err != nil {
			errstr := fmt.Sprint("errcode:%v", s.lasterror)
			err = errors.New(errstr) //s.lasterror
		}

		if s.SocketFd != INVALID_SOCKET {

			session.Close()
		}
		FinsReset(s)
	}

	return cliInfo, nil
}

func TcpErrorcodeToFinsRetval(errorcode uint32) int {

	switch errorcode {

	case 0x00000000:
		return FINS_RETVAL_CLOSED_BY_REMOTE
	case 0x00000001:
		return FINS_RETVAL_NO_FINS_HEADER
	case 0x00000002:
		return FINS_RETVAL_DATA_LENGTH_TOO_LONG
	case 0x00000003:
		return FINS_RETVAL_COMMAND_NOT_SUPPORTED
	case 0x00000020:
		return FINS_RETVAL_ALL_CONNECTIONS_IN_USE
	case 0x00000021:
		return FINS_RETVAL_NODE_ALREADY_CONNECTED
	case 0x00000022:
		return FINS_RETVAL_NODE_IP_PROTECTED
	case 0x00000023:
		return FINS_RETVAL_CLIENT_NODE_OUT_OF_RANGE
	case 0x00000024:
		return FINS_RETVAL_SAME_NODE_ADDRESS
	case 0x00000025:
		return FINS_RETVAL_NO_NODE_ADDRESS_AVAILABLE
	}

	return FINS_RETVAL_ILLEGAL_FINS_COMMAND
}

func FinsReset(sys *FinsSysTp) {

	if sys == nil {
		return
	}
	sys.SocketFd = INVALID_SOCKET
	sys.Timeout = time.Now().Unix()

	return
}

func XX_finslib_communicate(sys *FinsSysTp, command *fins_command_tp, bodylen *uint32) int {

	var a uint32
	var recvlen int
	var retval uint32
	var error_val uint32
	var endcode uint32
	var sent_header []uint8 = make([]byte, FINS_HEADER_LEN)
	var waste_buffer []uint8 = make([]byte, BUFLEN)

	if sys == nil {
		return check_error_count(sys, FINS_RETVAL_NOT_INITIALIZED)
	}
	if command == nil {
		return check_error_count(sys, FINS_RETVAL_NO_COMMAND)
	}
	if bodylen == nil {
		return check_error_count(sys, FINS_RETVAL_NO_COMMAND_LENGTH)
	}
	if sys.SocketFd == INVALID_SOCKET {
		return check_error_count(sys, FINS_RETVAL_NOT_CONNECTED)
	}

	if sys.CommType == FINS_COMM_TYPE_TCP {

		error_val = FINS_RETVAL_SUCCESS

		for a = 0; a < FINS_HEADER_LEN; a++ {
			sent_header[a] = command.header[a]
		}
		retval = fins_send_tcp_header(sys, *bodylen)
		if retval != FINS_RETVAL_SUCCESS {
			return check_error_count(sys, retval)
		}
		retval = fins_send_tcp_command(sys, *bodylen, command)
		if retval != FINS_RETVAL_SUCCESS {
			return check_error_count(sys, retval)
		}

		recvlen = fins_recv_tcp_header(sys, &error_val)

		if recvlen < 0 {
			return check_error_count(sys, error_val)
		}
		if recvlen == 0 {
			return check_error_count(sys, FINS_RETVAL_BODY_TOO_SHORT)
		}
		retval = fins_recv_tcp_command(sys, recvlen, command)
		if retval != FINS_RETVAL_SUCCESS {
			return check_error_count(sys, retval)
		}

		if command.header[FINS_ICF] != (sent_header[FINS_ICF]|0x40) ||
			command.header[FINS_RSV] != 0x00 ||
			command.header[FINS_DNA] != sent_header[FINS_SNA] ||
			command.header[FINS_DA1] != sent_header[FINS_SA1] ||
			command.header[FINS_DA2] != sent_header[FINS_SA2] ||
			command.header[FINS_SNA] != sent_header[FINS_DNA] ||
			command.header[FINS_SA1] != sent_header[FINS_DA1] ||
			command.header[FINS_SA2] != sent_header[FINS_DA2] ||
			command.header[FINS_SID] != sent_header[FINS_SID] ||
			command.header[FINS_MRC] != sent_header[FINS_MRC] ||
			command.header[FINS_SRC] != sent_header[FINS_SRC] {

			for {
				if fins_tcp_recv(sys, &waste_buffer, BUFLEN) <= 0 {
					break
				}
			}

			return check_error_count(sys, FINS_RETVAL_SYNC_ERROR)
		}

		recvlen -= FINS_HEADER_LEN
		*bodylen = uint32(recvlen)

		if recvlen < 2 {
			return check_error_count(sys, FINS_RETVAL_BODY_TOO_SHORT)
		}

		endcode = command.body[0] & 0x7f
		endcode <<= 8
		endcode += command.body[1] & 0x3f

		return check_error_count(sys, endcode)
	}

	return check_error_count(sys, FINS_RETVAL_NOT_INITIALIZED)

} /* XX_finslib_communicate */

func check_error_count(sys *FinsSysTp, error_code uint32) int {

	if sys == nil {
		return FINS_RETVAL_NOT_INITIALIZED
	}

	if sys.SocketFd == INVALID_SOCKET || sys.errorMax < 0 || error_code == FINS_RETVAL_SUCCESS || error_code == FINS_RETVAL_SUCCESS_LAST_DATA {
		sys.errorCount = 0
		sys.errorchange = (int(error_code) != sys.lasterror)
		sys.lasterror = int(error_code)
		return int(error_code)
	}

	sys.errorCount++

	if sys.errorCount > sys.errorMax {
		error_code = FINS_RETVAL_MAX_ERROR_COUNT
	}

	switch error_code {

	case FINS_RETVAL_MAX_ERROR_COUNT:
	case FINS_RETVAL_CLOSED_BY_REMOTE:
	case FINS_RETVAL_WSA_UNRECOGNIZED_ERROR:
	case FINS_RETVAL_WSA_NOT_INITIALIZED:
	case FINS_RETVAL_WSA_SYS_NOT_READY:
	case FINS_RETVAL_WSA_VER_NOT_SUPPORTED:
	case FINS_RETVAL_WSA_E_ACCES:
	case FINS_RETVAL_WSA_E_ADDR_IN_USE:
	case FINS_RETVAL_WSA_E_ADDR_NOT_AVAIL:
	case FINS_RETVAL_WSA_E_AF_NO_SUPPORT:
	case FINS_RETVAL_WSA_E_CONN_REFUSED:
	case FINS_RETVAL_WSA_E_HOST_UNREACH:
	case FINS_RETVAL_WSA_E_MFILE:
	case FINS_RETVAL_WSA_E_NET_DOWN:
	case FINS_RETVAL_WSA_E_NET_RESET:
	case FINS_RETVAL_WSA_E_NET_UNREACH:
	case FINS_RETVAL_WSA_E_NO_PROTO_OPT:
	case FINS_RETVAL_WSA_E_NOT_CONN:
	case FINS_RETVAL_WSA_E_NOT_SOCK:
	case FINS_RETVAL_WSA_E_PROC_LIM:
	case FINS_RETVAL_WSA_E_PROTO_NO_SUPPORT:
	case FINS_RETVAL_WSA_E_PROTO_TYPE:
	case FINS_RETVAL_WSA_E_PROVIDER_FAILED_INIT:
	case FINS_RETVAL_WSA_E_SOCKT_NO_SUPPORT:

		sys.errorCount = 0
		sys.errorchange = (int(error_code) != sys.lasterror)
		sys.lasterror = int(error_code)

		//fins_close_socket( sys )
		if sys.Session != nil {
			sys.Session.Close()
			sys.Session = nil
		}

		break
	}

	return int(error_code)

} /* check_error_count */

func fins_send_tcp_header(sys *FinsSysTp, bodylen uint32) uint32 {

	var sendlen int
	var fins_tcp_header [FINS_MAX_TCP_HEADER]uint32

	if sys == nil {
		return FINS_RETVAL_NOT_INITIALIZED
	}
	if sys.SocketFd == INVALID_SOCKET {
		return FINS_RETVAL_NOT_CONNECTED
	}
	if bodylen > FINS_BODY_LEN {
		return FINS_RETVAL_BODY_TOO_LONG
	}

	bodylen += 8 + FINS_HEADER_LEN

	fins_tcp_header[0] = 'F'
	fins_tcp_header[1] = 'I'
	fins_tcp_header[2] = 'N'
	fins_tcp_header[3] = 'S'

	fins_tcp_header[4] = (bodylen >> 24) & 0xff
	fins_tcp_header[5] = (bodylen >> 16) & 0xff
	fins_tcp_header[6] = (bodylen >> 8) & 0xff
	fins_tcp_header[7] = (bodylen) & 0xff

	fins_tcp_header[8] = 0x00
	fins_tcp_header[9] = 0x00
	fins_tcp_header[10] = 0x00
	fins_tcp_header[11] = 0x02

	fins_tcp_header[12] = 0x00
	fins_tcp_header[13] = 0x00
	fins_tcp_header[14] = 0x00
	fins_tcp_header[15] = 0x00

	sendlen = 16

	if send(sys.SocketFd, fins_tcp_header, sendlen, 0) != sendlen {
		return FINS_RETVAL_HEADER_SEND_ERROR
	}

	return FINS_RETVAL_SUCCESS

}

func fins_send_tcp_command(sys *FinsSysTp, bodylen uint32, command *fins_command_tp) uint32 {

	var sendlen uint32
	var retval uint32

	if sys == nil {
		return FINS_RETVAL_NOT_INITIALIZED
	}
	if command == nil {
		return FINS_RETVAL_NO_COMMAND
	}
	if sys.SocketFd == INVALID_SOCKET {
		return FINS_RETVAL_NOT_CONNECTED
	}
	if bodylen > FINS_BODY_LEN {
		return FINS_RETVAL_BODY_TOO_LONG
	}

	sendlen = FINS_HEADER_LEN + bodylen
	retval = send(sys.SocketFd, command, sendlen, 0)

	if retval < 0 {
		return FINS_RETVAL_ERRNO_BASE
	}
	if retval != sendlen {
		return FINS_RETVAL_COMMAND_SEND_ERROR
	}

	return FINS_RETVAL_SUCCESS
}

func fins_recv_tcp_header(sys *FinsSysTp, error_val *uint32) int {

	var recvlen uint32
	var retval uint32
	var command uint8
	var errorcode uint8
	var fins_tcp_header []byte = make([]byte, FINS_MAX_TCP_HEADER)

	if sys == nil || sys.SocketFd == INVALID_SOCKET {
		return -1
	}

	recvlen = 16
	retval = fins_tcp_recv(sys, &fins_tcp_header, recvlen)

	if retval < recvlen {
		if error_val != nil {
			*error_val = FINS_RETVAL_RESPONSE_HEADER_INCOMPLETE
		}
		return -1
	}

	command = fins_tcp_header[8]
	command <<= 8
	command += fins_tcp_header[9]
	command <<= 8
	command += fins_tcp_header[10]
	command <<= 8
	command += fins_tcp_header[11]

	errorcode = fins_tcp_header[12]
	errorcode <<= 8
	errorcode += fins_tcp_header[13]
	errorcode <<= 8
	errorcode += fins_tcp_header[14]
	errorcode <<= 8
	errorcode += fins_tcp_header[15]

	if command != 0x00000002 {

		if error_val != nil {
			*error_val = tcp_errorcode_to_fins_retval(errorcode)
		}
		return -1
	}

	recvlen = uint32(fins_tcp_header[6])
	recvlen <<= 8
	recvlen += uint32(fins_tcp_header[7])
	recvlen -= 8

	if recvlen > FINS_HEADER_LEN+FINS_BODY_LEN {
		if error_val != nil {
			*error_val = FINS_RETVAL_BODY_TOO_LONG
		}
		return -1
	}

	if error_val != nil {
		*error_val = FINS_RETVAL_SUCCESS
	}

	return int(recvlen)

}

func fins_recv_tcp_command(sys *FinsSysTp, total_len int, command *fins_command_tp) uint32 {

	if fins_tcp_recv(sys, []byte(command), uint32(total_len)) != uint32(total_len) {
		return FINS_RETVAL_RESPONSE_INCOMPLETE
	}
	return FINS_RETVAL_SUCCESS

} /* fins_recv_tcp_header */

func fins_tcp_recv(sys *FinsSysTp, buf *[]byte, len uint32) uint32 {

	var total_len uint32
	var recv_len uint32

	if len <= 0 {
		return 0
	}
	total_len = 0

	for {
		//ecv_len = recv(sys.SocketFd, buf, len, 0)
		sys.Session.Read()

		if recv_len > 0 {

			len -= recv_len
			*buf = append(*buf, uint8(recv_len))
			total_len += recv_len

			if len <= 0 {
				break
			}
		} else if recv_len < 0 {

		} else {
			return total_len
		}
	}

	return total_len

} /* fins_tcp_recv */
