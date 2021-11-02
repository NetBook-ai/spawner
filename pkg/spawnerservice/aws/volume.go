package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"go.uber.org/zap"
)

func LogError(methodName string, logger *zap.SugaredLogger, err error) {

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				logger.Errorw("Error in ", "method : ", methodName, "error", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logger.Errorw("Error in ", "method : ", methodName, "error", err.Error())
		}
		logger.Errorw("Error in ", "method : ", methodName, "error", err)
	}
}

func (svc AWSController) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	//Creates an EBS volume

	logger := svc.logger

	availabilityZone := req.GetAvailabilityzone()
	volumeType := req.GetVolumetype()
	size := req.GetSize()
	snapshotId := req.GetSnapshotid()
	region := req.Region

	input := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(availabilityZone),
		VolumeType:       aws.String(volumeType),
		Size:             aws.Int64(size),
		SnapshotId:       aws.String(snapshotId),
	}

	//creating session
	awsSvc, err := svc.sessionClient(region, logger)

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}

	//calling aws sdk CreateVolume function
	result, err := awsSvc.CreateVolume(input)

	if err != nil {
		LogError("CreateVolume", logger, err)
		return &pb.CreateVolumeResponse{}, err
	}

	res := &pb.CreateVolumeResponse{
		Volumeid: *result.VolumeId,
		Error:    "",
	}

	return res, nil
}

func (svc AWSController) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	//Deletes an EBS volume

	logger := svc.logger

	volumeid := req.GetVolumeid()
	region := req.Region

	input := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	//creating session
	awsSvc, err := svc.sessionClient(region, logger)

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}
	//calling aws sdk method to delete volume
	//ec2.DeleteVolumeOutput doesn't contain anything
	//hence not taking response
	_, err = awsSvc.DeleteVolume(input)

	if err != nil {
		LogError("DeleteVolume", logger, err)
		return &pb.DeleteVolumeResponse{}, err
	}

	//note: since now err is nil so assigning deleted = true
	res := &pb.DeleteVolumeResponse{
		Deleted: true,
	}

	return res, nil
}

func (svc AWSController) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	//Creates a Snapshot of a volume

	logger := svc.logger

	volumeid := req.GetVolumeid()
	region := req.Region

	input := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeid),
	}

	//creating session
	awsSvc, err := svc.sessionClient(region, logger)

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}

	//calling aws sdk method to snapshot volume
	result, err := awsSvc.CreateSnapshot(input)

	if err != nil {
		LogError("CreateSnapshot", logger, err)
		return &pb.CreateSnapshotResponse{}, err
	}

	res := &pb.CreateSnapshotResponse{
		Snapshotid: *result.SnapshotId,
	}

	return res, nil
}

func (svc AWSController) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
	//First Creates the snapshot of the volume then deletes the volume

	logger := svc.logger

	volumeid := req.GetVolumeid()
	region := req.Region

	inputSnapshot := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeid),
	}

	//creating session
	awsSvc, err := svc.sessionClient(region, logger)

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}

	//calling aws sdk CreateSnapshot method
	resultSnapshot, err := awsSvc.CreateSnapshot(inputSnapshot)

	if err != nil {
		LogError("CreateSnapshot", logger, err)
		return &pb.CreateSnapshotAndDeleteResponse{}, err
	}

	//inputs for deleteing volume
	inputDelete := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	//calling aws sdk method to delete volume
	//ec2.DeleteVolumeOutput doesn't contain anything
	//hence not taking response
	_, err = awsSvc.DeleteVolume(inputDelete)

	if err != nil {
		LogError("DeleteVolume", logger, err)

		return &pb.CreateSnapshotAndDeleteResponse{
			Snapshotid: *resultSnapshot.SnapshotId,
			Deleted:    false,
		}, err
	}

	//note: since now err is nil so assigning deleted = true
	res := &pb.CreateSnapshotAndDeleteResponse{
		Snapshotid: *resultSnapshot.SnapshotId,
		Deleted:    true,
	}

	return res, nil
}
