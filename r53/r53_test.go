package r53

import (
	"context"
	awsclient2 "ddns/awsclient"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/jcmturner/ddns/awsclient"
	"github.com/stretchr/testify/assert"
)

func TestUpdateRecord(t *testing.T) {
	r53Cl := new(awsclient2.R53Client)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := UpdateRecord(ctx, r53Cl, aws.String(awsclient.MockHostedZoneID), awsclient.MockeZoneName, "1.2.3.4", types.RRTypeA)
	if err != nil {
		t.Fatalf("error updating record: %v", err)
	}
}

func TestZoneID(t *testing.T) {
	r53Cl := new(awsclient2.R53Client)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	zid, err := ZoneID(ctx, r53Cl, awsclient.MockeZoneName)
	if err != nil {
		t.Fatalf("error getting zone id: %v", err)
	}
	assert.Equal(t, awsclient.MockHostedZoneID, aws.ToString(zid), "zone ID not expected")
}
