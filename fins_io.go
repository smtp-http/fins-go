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
