package sessiongrpc

import (
	sessionproto "LTICore/api/session/v1"
	"LTICore/internal/core/service"
	"context"
	"fmt"
	"sync"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a dynamic gRPC client for session.v1.SessionService (proto source: api/session/v1).
type Client struct {
	conn          *grpc.ClientConn
	stub          grpcdynamic.Stub
	createMethod  *desc.MethodDescriptor
	reqMsgType    *desc.MessageDescriptor
	launchTypeLTI int32
	parseErr      error
	once          sync.Once
}

// New dials session_service over plaintext gRPC (add TLS separately for production backends).
func New(target string) (*Client, error) {
	if target == "" {
		return nil, fmt.Errorf("session grpc target is empty")
	}
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("session grpc dial: %w", err)
	}

	c := &Client{
		conn: conn,
		stub: grpcdynamic.NewStub(conn),
	}
	if err := c.ensureDescriptors(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return c, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) ensureDescriptors() error {
	c.once.Do(func() {
		accessor := protoparse.FileContentsFromMap(map[string]string{
			"session.proto":                   string(sessionproto.SessionProto),
			"google/protobuf/timestamp.proto": string(sessionproto.TimestampProto),
		})
		parser := protoparse.Parser{Accessor: accessor}
		fds, err := parser.ParseFiles("session.proto")
		if err != nil {
			c.parseErr = err
			return
		}
		var svcDesc *desc.ServiceDescriptor
		for _, fd := range fds {
			sym := fd.FindSymbol("session.v1.SessionService")
			if sym == nil {
				continue
			}
			var ok bool
			svcDesc, ok = sym.(*desc.ServiceDescriptor)
			if ok {
				break
			}
		}
		if svcDesc == nil {
			c.parseErr = fmt.Errorf("descriptor session.v1.SessionService not found")
			return
		}
		c.createMethod = svcDesc.FindMethodByName("CreateSession")
		if c.createMethod == nil {
			c.parseErr = fmt.Errorf("descriptor CreateSession not found")
			return
		}
		c.reqMsgType = c.createMethod.GetInputType()
		ltf := c.reqMsgType.FindFieldByName("launch_type")
		if ltf == nil || ltf.GetEnumType() == nil {
			c.parseErr = fmt.Errorf("launch_type field missing")
			return
		}
		ev := ltf.GetEnumType().FindValueByName("LAUNCH_TYPE_LTI")
		if ev == nil {
			c.parseErr = fmt.Errorf("enum LAUNCH_TYPE_LTI missing")
			return
		}
		c.launchTypeLTI = int32(ev.GetNumber())
	})
	return c.parseErr
}

// CreateForLTILaunch implements service.EducationSessionCreator.
func (c *Client) CreateForLTILaunch(ctx context.Context, p service.EducationSessionCreateParams) (string, error) {
	if err := c.ensureDescriptors(); err != nil {
		return "", err
	}
	req := dynamic.NewMessage(c.reqMsgType)
	_ = req.TrySetFieldByName("request_id", p.RequestID)
	_ = req.TrySetFieldByName("user_id", p.UserID)
	_ = req.TrySetFieldByName("course_id", p.CourseID)
	_ = req.TrySetFieldByName("launch_type", c.launchTypeLTI)
	if p.ResourceID != "" {
		_ = req.TrySetFieldByName("resource_id", p.ResourceID)
	}

	resp, err := c.stub.InvokeRpc(ctx, c.createMethod, req)
	if err != nil {
		return "", err
	}
	dm, ok := resp.(*dynamic.Message)
	if !ok {
		return "", fmt.Errorf("unexpected response type %T", resp)
	}
	sessVal, err := dm.TryGetFieldByName("session")
	if err != nil || sessVal == nil {
		return "", fmt.Errorf("create session response missing session")
	}
	sess, ok := sessVal.(*dynamic.Message)
	if !ok {
		return "", fmt.Errorf("session field has wrong type %T", sessVal)
	}
	idVal, err := sess.TryGetFieldByName("id")
	if err != nil {
		return "", err
	}
	sid, ok := idVal.(string)
	if !ok {
		return "", fmt.Errorf("session.id is not a string")
	}
	return sid, nil
}
