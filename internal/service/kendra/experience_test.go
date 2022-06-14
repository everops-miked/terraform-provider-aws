package kendra_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestExperienceParseResourceID(t *testing.T) {
	testCases := []struct {
		TestName        string
		Input           string
		ExpectedId      string
		ExpectedIndexId string
		Error           bool
	}{
		{
			TestName:        "empty",
			Input:           "",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID",
			Input:           "abcdefg12345678/",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID separator",
			Input:           "abcdefg12345678:qwerty09876",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID with more than 1 separator",
			Input:           "abcdefg12345678/qwerty09876/zxcvbnm123456",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Valid ID",
			Input:           "abcdefg12345678/qwerty09876",
			ExpectedId:      "abcdefg12345678",
			ExpectedIndexId: "qwerty09876",
			Error:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotId, gotIndexId, err := tfkendra.ExperienceParseResourceID(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (Id: %s, IndexId: %s) and no error, expected error", gotId, gotIndexId)
			}

			if gotId != testCase.ExpectedId {
				t.Errorf("got %s, expected %s", gotId, testCase.ExpectedIndexId)
			}

			if gotIndexId != testCase.ExpectedIndexId {
				t.Errorf("got %s, expected %s", gotIndexId, testCase.ExpectedIndexId)
			}
		})
	}
}

func TestAccKendraExperience_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`experience/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "experience_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKendraExperience_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfkendra.ResourceExperience(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKendraExperience_description(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"
	description := "Example Description"
	descriptionUpdated := fmt.Sprintf("Updated %s", description)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`experience/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "experience_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExperienceConfig_description(rName, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`experience/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "experience_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`experience/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "experience_id"),
				),
			},
		},
	})
}

func TestAccKendraExperience_name(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccExperienceConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKendraExperience_Configuration_ContentSourceConfiguration_DirectPutContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_Configuration_ContentSourceConfiguration_DirectPutContent(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExperienceConfig_Configuration_ContentSourceConfiguration_DirectPutContent(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExperienceConfig_Configuration_ContentSourceConfiguration_DirectPutContent(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "true"),
				),
			},
		},
	})
}

func TestAccKendraExperience_Configuration_UserIdentityConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"
	userId := "92677ec453-1546651e-79c4-4554-91fa-00b43ccfa245"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
			testAccPreCheckIdentityStoreUser(t, userId)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_Configuration_UserIdentityConfiguration(rName, userId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.0.identity_attribute_name", userId),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckExperienceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kendra_experience" {
			continue
		}

		id, indexId, err := tfkendra.ExperienceParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &kendra.DescribeExperienceInput{
			Id:      aws.String(id),
			IndexId: aws.String(indexId),
		}

		_, err = conn.DescribeExperience(context.TODO(), input)

		var resourceNotFoundException *types.ResourceNotFoundException

		if errors.As(err, &resourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Expected Kendra Experience to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckExperienceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kendra Experience is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

		id, indexId, err := tfkendra.ExperienceParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.DescribeExperience(context.TODO(), &kendra.DescribeExperienceInput{
			Id:      aws.String(id),
			IndexId: aws.String(indexId),
		})

		if err != nil {
			return fmt.Errorf("Error describing Kendra Experience: %s", err.Error())
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

	input := &kendra.ListIndicesInput{}

	_, err := conn.ListIndices(context.TODO(), input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPreCheckIdentityStoreUser(t *testing.T, userId string) {
	ssoAdminConn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn
	identityStoreConn := acctest.Provider.Meta().(*conns.AWSClient).IdentityStoreConn

	input := &ssoadmin.ListInstancesInput{
		MaxResults: aws.Int64(1),
	}

	output, err := ssoAdminConn.ListInstancesWithContext(context.TODO(), input)
	if err != nil {
		t.Fatalf("error listing SSO Admin instances: %s", err)
	}

	if output == nil || len(output.Instances) == 0 {
		t.Skip("skipping acceptance testing: no SSO Admin instances")
	}

	listUsersInput := &identitystore.DescribeUserInput{
		IdentityStoreId: output.Instances[0].IdentityStoreId,
		UserId:          aws.String(userId),
	}

	resp, err := identityStoreConn.DescribeUserWithContext(context.TODO(), listUsersInput)
	if err != nil {
		t.Fatalf("error describing IdentityStore User (%s): %s", userId, err)
	}

	if resp == nil {
		t.Skipf("skipping acceptance testing: no matching IdentityStore user with userID %s", userId)
	}
}

func testAccExperienceBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      type        = "Service"
      identifiers = ["kendra.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData"
    ]
    resources = [
      "*"
    ]
    condition {
      test     = "StringEquals"
      variable = "cloudwatch:namespace"
      values   = ["Kendra"]
    }
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogGroups"
    ]
    resources = [
      "*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogStreams",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*:log-stream:*"
    ]
  }
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  description = "Allow Kendra to access cloudwatch logs"
  policy      = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "experience" {
  statement {
    sid    = "AllowsKendraSearchAppToCallKendraApi"
    effect = "Allow"
    actions = [
      "kendra:GetQuerySuggestions",
      "kendra:Query",
      "kendra:DescribeIndex",
      "kendra:ListFaqs",
      "kendra:DescribeDataSource",
      "kendra:ListDataSources",
      "kendra:DescribeFaq"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:kendra:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:index/${aws_kendra_index.test.id}"
    ]
  }
}

resource "aws_iam_policy" "experience" {
  name        = "%[1]s-experience"
  description = "Allow Kendra to search app access"
  policy      = data.aws_iam_policy_document.experience.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_iam_role_policy_attachment" "experience" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.experience.arn
}

resource "aws_kendra_index" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccExperienceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
}
`, rName))
}

func testAccExperienceConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id    = aws_kendra_index.test.id
  description = %[2]q
  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
}
`, rName, description))
}

func testAccExperienceConfig_Configuration_ContentSourceConfiguration_DirectPutContent(rName string, directPutContent bool) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  configuration {
    content_source_configuration {
      direct_put_content = %[2]t
    }
  }
}
`, rName, directPutContent))
}

func testAccExperienceConfig_Configuration_UserIdentityConfiguration(rName, userId string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  configuration {
    user_identity_configuration {
      identity_attribute_name = %[2]q
    }
  }
}
`, rName, userId))
}
