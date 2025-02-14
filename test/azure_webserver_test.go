package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/gruntwork-io/terratest/modules/ssh"
)

// You normally want to run this under a separate "Testing" subscription
// For lab purposes you will use your assigned subscription under the Cloud Dev/Ops program tenant
var subscriptionID string = "431fca8d-e614-4268-aa3c-22a2e684933a"

func TestAzureLinuxVMCreation(t *testing.T) {
	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../",
		// Override the default terraform variables
		Vars: map[string]interface{}{
			"labelPrefix": "gajj0008",
		},
	}

	defer terraform.Destroy(t, terraformOptions)

	// Run `terraform init` and `terraform apply`. Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the value of output variables
	vmName := terraform.Output(t, terraformOptions, "vm_name")
	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")
	nicName := terraform.Output(t, terraformOptions, "nic_name") // Assuming you have a NIC output
	publicIP := terraform.Output(t, terraformOptions, "public_ip")

	// Confirm VM exists
	assert.True(t, azure.VirtualMachineExists(t, vmName, resourceGroupName, subscriptionID))

	// Confirm NIC exists and is connected to the VM
	assert.True(t, azure.NetworkInterfaceExists(t, nicName, resourceGroupName, subscriptionID))

	// Fetch NIC details and verify if it is connected to the VM
	nic, err := azure.GetNetworkInterface(t, nicName, resourceGroupName, subscriptionID)
	assert.NoError(t, err)
	assert.Equal(t, nic.VirtualMachineID, vmName)

	// Confirm the VM is running the correct Ubuntu version (for example, Ubuntu 22.04)
	// Connect via SSH to the VM and run `lsb_release -a` to confirm the Ubuntu version
	sshKeyPath := "~/.ssh/id_rsa" // Replace with your SSH key path
	sshHost := publicIP
	sshConfig := &ssh.SSHConfig{
		User: "azureuser", // Update with your actual username
		PrivateKey: ssh.KeyPair{Key: sshKeyPath},
	}
	sshOptions := &ssh.ClientOptions{
		Config: sshConfig,
	}
	// SSH into the VM and run the command to check the Ubuntu version
	ubuntuVersionCmd := "lsb_release -a"
	sshResult := ssh.RunCommand(t, sshOptions, sshHost, ubuntuVersionCmd)

	// Ensure the output contains the expected Ubuntu version
	assert.Contains(t, sshResult.Stdout, "Ubuntu 22.04")

	// Alternatively, you can check the `uname -a` for kernel info if needed
	// kernelInfoCmd := "uname -a"
	// kernelInfoResult := ssh.RunCommand(t, sshOptions, sshHost, kernelInfoCmd)
	// assert.Contains(t, kernelInfoResult.Stdout, "Ubuntu 22.04")
}
