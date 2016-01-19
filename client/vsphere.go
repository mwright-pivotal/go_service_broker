package client

import (
	"net/url"
        "os"
        "fmt"
        "strconv"
        
        "golang.org/x/net/context"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"	
//	"github.com/vmware/govmomi"
)

const (
        envURL      = "GOVMOMI_URL"
        envUserName = "GOVMOMI_USERNAME"
        envPassword = "GOVMOMI_PASSWORD"
        envInsecure = "GOVMOMI_INSECURE"
)

type VmwareClient struct {
        vcenterClient *vim25.Client
}

func NewVmwareClient() *VmwareClient {
        fmt.Println("NewVmwareClient ready!")

        return &VmwareClient{
                vcenterClient: NewAuthenticatedClient(),         
        }       
}

// state == pending, running, succeeded, failed
func (c *VmwareClient) GetInstanceState(instanceId string) (string, error) {
        vgId, err := strconv.Atoi(instanceId)
        if err != nil {
                return "failed", err
        }

	return strconv.Itoa(vgId), nil
}

func (c *VmwareClient) CreateInstance(parameters interface{}) (string, error) {
	return "", nil
}

func (c *VmwareClient) InjectKeyPair(instanceId string) (string, string, string, error) {
        return "", "", "", nil
}

func (c *VmwareClient) DeleteInstance(instanceId string) error {
        vgId, err := strconv.Atoi(instanceId)
        if err != nil {
                return err
        }
	
	strconv.Itoa(vgId)
	return  nil
}

func (c *VmwareClient) RevokeKeyPair(instanceId string, privateKeyName string) error {
	return nil
}

// URL parses the GOVMOMI_TEST_URL environment variable if set.
func URL() *url.URL {
	s := os.Getenv("GOVC_URL")
	if s == "" {
		return nil
	}
	u, err := soap.ParseURL(s)
	if err != nil {
		panic(err)
	}
	return u
}

// NewAuthenticatedClient creates a new vim25.Client, authenticates the user
// specified in the test URL, and returns it.
func NewAuthenticatedClient() *vim25.Client {
	u := URL()

	soapClient := soap.NewClient(u, true)
	vimClient, err := vim25.NewClient(context.Background(), soapClient)
	if err != nil {
		panic(err)
	}

	req := types.Login{
		This: *vimClient.ServiceContent.SessionManager,
	}

	req.UserName = u.User.Username()
	if pw, ok := u.User.Password(); ok {
		req.Password = pw
	}

	_, err = methods.Login(context.Background(), vimClient, &req)
	if err != nil {
		panic(err)
	}

	return vimClient
}
