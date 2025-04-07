#!/bin/bash

# Configuration
REGION="us-east-1"
URL_TABLE="url-shortener"
COUNTER_TABLE="url-counter"

# Minimum and maximum capacity units
MIN_RCU=5
MAX_RCU=100
MIN_WCU=5
MAX_WCU=100

# Target utilization percentage (70% is recommended)
TARGET_UTILIZATION=70

# Create auto-scaling targets for URL table
echo "Setting up auto-scaling for URL table..."

# Read capacity
aws application-autoscaling register-scalable-target \
    --service-namespace dynamodb \
    --scalable-dimension dynamodb:table:ReadCapacityUnits \
    --resource-id "table/$URL_TABLE" \
    --min-capacity $MIN_RCU \
    --max-capacity $MAX_RCU \
    --region $REGION

# Write capacity
aws application-autoscaling register-scalable-target \
    --service-namespace dynamodb \
    --scalable-dimension dynamodb:table:WriteCapacityUnits \
    --resource-id "table/$URL_TABLE" \
    --min-capacity $MIN_WCU \
    --max-capacity $MAX_WCU \
    --region $REGION

# Create auto-scaling policies for URL table
aws application-autoscaling put-scaling-policy \
    --policy-name "${URL_TABLE}-read-scaling-policy" \
    --service-namespace dynamodb \
    --scalable-dimension dynamodb:table:ReadCapacityUnits \
    --resource-id "table/$URL_TABLE" \
    --policy-type TargetTrackingScaling \
    --target-tracking-scaling-policy-configuration "{
        \"TargetValue\": $TARGET_UTILIZATION,
        \"PredefinedMetricSpecification\": {
            \"PredefinedMetricType\": \"DynamoDBReadCapacityUtilization\"
        },
        \"ScaleOutCooldown\": 60,
        \"ScaleInCooldown\": 60
    }" \
    --region $REGION

aws application-autoscaling put-scaling-policy \
    --policy-name "${URL_TABLE}-write-scaling-policy" \
    --service-namespace dynamodb \
    --scalable-dimension dynamodb:table:WriteCapacityUnits \
    --resource-id "table/$URL_TABLE" \
    --policy-type TargetTrackingScaling \
    --target-tracking-scaling-policy-configuration "{
        \"TargetValue\": $TARGET_UTILIZATION,
        \"PredefinedMetricSpecification\": {
            \"PredefinedMetricType\": \"DynamoDBWriteCapacityUtilization\"
        },
        \"ScaleOutCooldown\": 60,
        \"ScaleInCooldown\": 60
    }" \
    --region $REGION

# Create auto-scaling targets for Counter table
echo "Setting up auto-scaling for Counter table..."

# Read capacity
aws application-autoscaling register-scalable-target \
    --service-namespace dynamodb \
    --scalable-dimension dynamodb:table:ReadCapacityUnits \
    --resource-id "table/$COUNTER_TABLE" \
    --min-capacity $MIN_RCU \
    --max-capacity $MAX_RCU \
    --region $REGION

# Write capacity
aws application-autoscaling register-scalable-target \
    --service-namespace dynamodb \
    --scalable-dimension dynamodb:table:WriteCapacityUnits \
    --resource-id "table/$COUNTER_TABLE" \
    --min-capacity $MIN_WCU \
    --max-capacity $MAX_WCU \
    --region $REGION

# Create auto-scaling policies for Counter table
aws application-autoscaling put-scaling-policy \
    --policy-name "${COUNTER_TABLE}-read-scaling-policy" \
    --service-namespace dynamodb \
    --scalable-dimension dynamodb:table:ReadCapacityUnits \
    --resource-id "table/$COUNTER_TABLE" \
    --policy-type TargetTrackingScaling \
    --target-tracking-scaling-policy-configuration "{
        \"TargetValue\": $TARGET_UTILIZATION,
        \"PredefinedMetricSpecification\": {
            \"PredefinedMetricType\": \"DynamoDBReadCapacityUtilization\"
        },
        \"ScaleOutCooldown\": 60,
        \"ScaleInCooldown\": 60
    }" \
    --region $REGION

aws application-autoscaling put-scaling-policy \
    --policy-name "${COUNTER_TABLE}-write-scaling-policy" \
    --service-namespace dynamodb \
    --scalable-dimension dynamodb:table:WriteCapacityUnits \
    --resource-id "table/$COUNTER_TABLE" \
    --policy-type TargetTrackingScaling \
    --target-tracking-scaling-policy-configuration "{
        \"TargetValue\": $TARGET_UTILIZATION,
        \"PredefinedMetricSpecification\": {
            \"PredefinedMetricType\": \"DynamoDBWriteCapacityUtilization\"
        },
        \"ScaleOutCooldown\": 60,
        \"ScaleInCooldown\": 60
    }" \
    --region $REGION

echo "Auto-scaling setup complete!" 