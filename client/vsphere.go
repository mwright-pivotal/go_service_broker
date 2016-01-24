package client

import (
	"net/url"
        "os"
        "fmt"
        "strconv"
        
        "golang.org/x/net/context"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
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
	vFinder *find.Finder
	vDatacenter *object.Datacenter
	vgProps vmwareGuestProps
}

type vmwareGuestProps struct {
        hostname string
        domain string
        startCpus int
        maxMemory int
        dataCenterName string
	resourcePool string
	vmNetwork string
        operatingSystemReferenceCode string
}

func NewVmwareClient() *VmwareClient {
        fmt.Println("NewVmwareClient ready!")
	var c = NewAuthenticatedClient()
	var f = find.NewFinder(c, true)
	var dc, err = f.DefaultDatacenter(context.Background())
	if err != nil {
		fmt.Println("Failed to get default Datacenter")
		panic("help!")
	}

	f.SetDatacenter(dc)

	defaultProps := defaultVmwareGuestProperties()

        return &VmwareClient{
                vcenterClient: c,
		vFinder: f,
		vDatacenter: dc,
		vgProps: defaultProps,
        }       
}

// state == pending, running, succeeded, failed
func (c *VmwareClient) GetInstanceState(instanceId string) (string, error) {
//        vgId, err := strconv.Atoi(instanceId)
//        if err != nil {
//                return "failed", err
//        }

	return "running", nil
}

func (c *VmwareClient) CreateInstance(parameters interface{}) (string, error) {
	vm, err := c.vFinder.VirtualMachine(context.Background(), "base_ubuntu_template")
	if err != nil {
		return "failed", err
	}
	nm, err := vm.Name(context.Background())
	if err != nil {
		panic(err)
        }
	fmt.Println("found vm " + nm)
	
	newVm, err := c.CloneVM(vm, "changemyname")

	if err != nil {
                panic(err)
        }
	var o mo.VirtualMachine

	err = newVm.Properties(context.Background(), newVm.Reference(), []string{}, &o)
	if err != nil {
		return "", err
	}
	s := o.Summary

	return s.Config.Uuid, nil
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
	s := os.Getenv("VCENTER_URL")
	if s == "" {
		panic("VCENTER_URL env not set.  Please set to url of your vcenter sdk endpoint")
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

func defaultVmwareGuestProperties() vmwareGuestProps {
        return vmwareGuestProps {
                hostname: "go-service-broker",
                domain: "sys-wright.dyndns.ws",
                startCpus: 1,
                maxMemory: 1024,
                dataCenterName: "Wright Datacenter",
		resourcePool: "pcfPool",
		vmNetwork: "VM Network",
                operatingSystemReferenceCode: "base_ubuntu_template",
        }
}

//func getFlagSet(parameters interface{}) flag.FlagSet {
//
//}
func (c *VmwareClient) CloneVM(vm *object.VirtualMachine, name string) (*object.VirtualMachine, error) {
	folders, err := c.vDatacenter.Folders(context.TODO())
	if err != nil {
		return nil, err
	}

	spec := types.VirtualMachineCloneSpec{
		Config:   &types.VirtualMachineConfigSpec{},
		Location: types.VirtualMachineRelocateSpec{},
	}

	task, err := vm.Clone(context.TODO(), folders.VmFolder, name, spec)
	if err != nil {
		return nil, err
	}

	info, err := task.WaitForResult(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	return object.NewVirtualMachine(c.vcenterClient, info.Result.(types.ManagedObjectReference)), nil
}
