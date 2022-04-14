package base

import (
	"fmt"
	"github.com/derekyu332/goii/frame/rabbit"
	"github.com/derekyu332/goii/helper/encode"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/golang/protobuf/proto"
	"time"
)

const (
	NETHEAD_LENGTH       = 54
	CSHEAD_LENGTH        = 32
	NETHEAD_FLAG_CORR_ID = 8
)

type NetHead struct {
	Ver       uint16
	SrcFE     uint16
	SrcID     uint16
	DstFE     uint16
	DstID     uint16
	SocketFD  uint32
	SrcIP     uint32
	SrcPort   uint16
	Uin       uint32
	MsgID     uint32
	MsgType   uint16
	MsgSeq    uint32
	MsgBit    uint32
	TimeStamp uint32
	SessionID uint32
	Code      uint32
	Flag      uint32
}

func (this *NetHead) Encode() []byte {
	buf := make([]byte, 0)
	buf = encode.EncodeUint16(buf, this.Ver)
	buf = encode.EncodeUint16(buf, this.SrcFE)
	buf = encode.EncodeUint16(buf, this.SrcID)
	buf = encode.EncodeUint16(buf, this.DstFE)
	buf = encode.EncodeUint16(buf, this.DstID)
	buf = encode.EncodeUint32(buf, this.SocketFD)
	buf = encode.EncodeUint32(buf, this.SrcIP)
	buf = encode.EncodeUint16(buf, this.SrcPort)
	buf = encode.EncodeUint32(buf, this.Uin)
	buf = encode.EncodeUint32(buf, this.MsgID)
	buf = encode.EncodeUint16(buf, this.MsgType)
	buf = encode.EncodeUint32(buf, this.MsgSeq)
	buf = encode.EncodeUint32(buf, this.MsgBit)
	buf = encode.EncodeUint32(buf, this.TimeStamp)
	buf = encode.EncodeUint32(buf, this.SessionID)
	buf = encode.EncodeUint32(buf, this.Code)
	buf = encode.EncodeUint32(buf, this.Flag)

	return buf
}

func (this *NetHead) Decode(body []byte) []byte {
	buf := body
	buf, this.Ver = encode.DecodeUint16(buf)
	buf, this.SrcFE = encode.DecodeUint16(buf)
	buf, this.SrcID = encode.DecodeUint16(buf)
	buf, this.DstFE = encode.DecodeUint16(buf)
	buf, this.DstID = encode.DecodeUint16(buf)
	buf, this.SocketFD = encode.DecodeUint32(buf)
	buf, this.SrcIP = encode.DecodeUint32(buf)
	buf, this.SrcPort = encode.DecodeUint16(buf)
	buf, this.Uin = encode.DecodeUint32(buf)
	buf, this.MsgID = encode.DecodeUint32(buf)
	buf, this.MsgType = encode.DecodeUint16(buf)
	buf, this.MsgSeq = encode.DecodeUint32(buf)
	buf, this.MsgBit = encode.DecodeUint32(buf)
	buf, this.TimeStamp = encode.DecodeUint32(buf)
	buf, this.SessionID = encode.DecodeUint32(buf)
	buf, this.Code = encode.DecodeUint32(buf)
	buf, this.Flag = encode.DecodeUint32(buf)

	return buf
}

type CSHead struct {
	Ver       uint16
	Uin       uint32
	MsgID     uint32
	MsgSeq    uint32
	TimeStamp uint32
	SessionID uint32
	Code      uint32
}

func (this *CSHead) Encode() []byte {
	buf := make([]byte, 0)
	buf = encode.EncodeUint16(buf, this.Ver)
	buf = encode.EncodeUint32(buf, this.Uin)
	buf = encode.EncodeUint32(buf, this.MsgID)
	buf = encode.EncodeUint32(buf, this.MsgSeq)
	buf = encode.EncodeUint32(buf, this.TimeStamp)
	buf = encode.EncodeUint32(buf, this.SessionID)
	buf = encode.EncodeUint32(buf, this.Code)
	buf = encode.EncodeByte(buf, 0)
	buf = encode.EncodeByte(buf, 0)
	buf = encode.EncodeUint16(buf, 0)
	buf = encode.EncodeUint16(buf, 0)

	return buf
}

func (this *CSHead) Decode(body []byte) []byte {
	buf := body
	buf, this.Ver = encode.DecodeUint16(buf)
	buf, this.Uin = encode.DecodeUint32(buf)
	buf, this.MsgID = encode.DecodeUint32(buf)
	buf, this.MsgSeq = encode.DecodeUint32(buf)
	buf, this.TimeStamp = encode.DecodeUint32(buf)
	buf, this.SessionID = encode.DecodeUint32(buf)
	buf, this.Code = encode.DecodeUint32(buf)

	return buf[6:]
}

type ServerCall struct {
	RequestID int64
}

func (this *ServerCall) RpcServer(request proto.Message, netHead *NetHead, csHead *CSHead, response proto.Message) error {
	head := netHead.Encode()
	cs := csHead.Encode()
	body, err := proto.Marshal(request)

	if err != nil {
		logger.Warning("[%v] proto.Marshal failed %v", this.RequestID, err.Error())

		return err
	} else if len(body) >= 65535 {
		logger.Warning("[%v] Error: Too Long Body %v", this.RequestID, len(body))

		return ServerUnexpectedHttpError(nil, "Too Long Body")
	}

	msgLenByte := make([]byte, 0)
	msgLenByte = encode.EncodeUint16(msgLenByte, (uint16)(len(body)))
	code := encode.BytesCombine(head, cs, msgLenByte, body)
	key := fmt.Sprintf("F%dS%d", netHead.DstFE, netHead.DstID)
	logger.Info("[%v] Len = %v, Key = %v", this.RequestID, len(code), key)
	var rspByte []byte
	rspByte, err = rabbit.RPC(code, key, 3*time.Second)

	if err != nil {
		logger.Warning("[%v] rabbit.RPC failed %v", this.RequestID, err.Error())

		return err
	}

	protoBuf := rspByte[NETHEAD_LENGTH+CSHEAD_LENGTH+2:]
	err = proto.Unmarshal(protoBuf, response)

	if err != nil {
		logger.Warning("[%v] proto.Unmarshal failed %v", this.RequestID, err.Error())

		return err
	}

	return nil
}

func (this *ServerCall) NotifyServer(message proto.Message, netHead *NetHead, csHead *CSHead) error {
	head := netHead.Encode()
	cs := csHead.Encode()
	body, err := proto.Marshal(message)

	if err != nil {
		logger.Warning("[%v] proto.Marshal failed %v", this.RequestID, err.Error())

		return err
	} else if len(body) >= 65535 {
		logger.Warning("[%v] Error: Too Long Body %v", this.RequestID, len(body))

		return ServerUnexpectedHttpError(nil, "Too Long Body")
	}

	msgLenByte := make([]byte, 0)
	msgLenByte = encode.EncodeUint16(msgLenByte, (uint16)(len(body)))
	code := encode.BytesCombine(head, cs, msgLenByte, body)
	key := fmt.Sprintf("F%dS%d", netHead.DstFE, netHead.DstID)
	logger.Info("[%v] Len = %v, Key = %v", this.RequestID, len(code), key)
	err = rabbit.Notify(code, key)

	if err != nil {
		logger.Warning("[%v] rabbit.Notify failed %v", this.RequestID, err.Error())

		return err
	}

	return nil
}
