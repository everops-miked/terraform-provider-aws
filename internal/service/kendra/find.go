package kendra

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindQuerySuggestionsBlockListByID(ctx context.Context, conn *kendra.Client, id, indexId string) (*kendra.DescribeQuerySuggestionsBlockListOutput, error) {
	in := &kendra.DescribeQuerySuggestionsBlockListInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	}

	out, err := conn.DescribeQuerySuggestionsBlockList(ctx, in)
	if err != nil {
		var notFound *types.ResourceNotFoundException

		if errors.As(err, &notFound) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
