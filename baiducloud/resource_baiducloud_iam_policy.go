/*
Provide a resource to manage an IAM Policy.

Example Usage

```hcl
resource "baiducloud_iam_policy" "my-policy" {
  name = "my_policy"
  description = "my description"
  document = <<EOF
{"accessControlList": [{"region":"bj","service":"bcc","resource":["*"],"permission":["*"],"effect":"Allow"}]}
  EOF
}
```
*/
package baiducloud

import (
	"bytes"
	"encoding/json"
	"github.com/baidubce/bce-sdk-go/services/iam"
	"github.com/baidubce/bce-sdk-go/services/iam/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-baiducloud/baiducloud/connectivity"
	"time"
)

func resourceBaiduCloudIamPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceBaiduCloudIamPolicyCreate,
		Read:   resourceBaiduCloudIamPolicyRead,
		// TODO: sdk currently not support policy update
		// Update: resourceBaiduCloudIamPolicyUpdate,
		Delete: resourceBaiduCloudIamPolicyDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"unique_id": {
				Type:        schema.TypeString,
				Description: "Unique ID of policy.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of policy.",
				Required:    true,
				ForceNew:    true,
			},
			"document": {
				Type:        schema.TypeString,
				Description: "Json serialized ACL string.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// id, eid may be generated by remote, ignore when diff
					oldAcl := &api.Acl{}
					json.Unmarshal([]byte(old), &oldAcl)
					newAcl := &api.Acl{}
					json.Unmarshal([]byte(new), &newAcl)
					oldAcl.Id = ""
					newAcl.Id = ""
					for _, entry := range oldAcl.AccessControlList {
						entry.Eid = ""
					}
					for _, entry := range newAcl.AccessControlList {
						entry.Eid = ""
					}

					oldJson, _ := json.Marshal(oldAcl)
					newJson, _ := json.Marshal(newAcl)
					return bytes.Compare(oldJson, newJson) == 0
				},
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the policy.",
				Optional:    true,
				ForceNew:    true,
			},
			// TODO: support force_destroy
			/*
				"force_destroy": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Delete policy and its related user and group policy attachment.",
				},
			*/
		},
	}
}

func resourceBaiduCloudIamPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	document := d.Get("document").(string)
	action := "Create Policy " + name

	policy, err := client.WithIamClient(func(iamClient *iam.Client) (i interface{}, e error) {
		return iamClient.CreatePolicy(&api.CreatePolicyArgs{
			Name:        name,
			Document:    document,
			Description: description,
		})
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_iam_policy", action, BCESDKGoERROR)
	}
	addDebug(action, policy)

	d.SetId(name)
	return resourceBaiduCloudIamPolicyRead(d, meta)
}

func resourceBaiduCloudIamPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)
	name := d.Id()
	action := "Query Policy " + name

	raw, err := client.WithIamClient(func(iamClient *iam.Client) (i interface{}, e error) {
		return iamClient.GetPolicy(name, api.POLICY_TYPE_CUSTOM)
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_iam_policy", action, BCESDKGoERROR)
	}
	addDebug(action, raw)

	policy, _ := raw.(*api.GetPolicyResult)
	d.Set("unique_id", policy.Id)
	d.Set("name", policy.Name)
	d.Set("document", policy.Document)
	d.Set("description", policy.Description)
	return nil
}

func resourceBaiduCloudIamPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.BaiduClient)
	name := d.Id()
	action := "Delete Policy " + name

	_, err := client.WithIamClient(func(iamClient *iam.Client) (i interface{}, e error) {
		return nil, iamClient.DeletePolicy(name)
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "baiducloud_iam_policy", action, BCESDKGoERROR)
	}
	addDebug(action, nil)
	return nil
}
