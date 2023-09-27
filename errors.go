package rabbitmq

import (
        "errors"
)

var MessageNotSent error = errors.New("Message failed to sent")
var MessagesLimitErr error = errors.New("Message failed to sent")

var ErrorMissedConnectionConfig  error = errors.New("Missed rabbitmq connection config")
var ErrorMissedExchangeConfig    error = errors.New("Missed rabbitmq exchange config")
var ErrorMissedQueueConfig       error = errors.New("Missed rabbitmq queue config")
var ErrorMissedBindingConfig     error = errors.New("Missed rabbitmq binding config")

var ErrorMissedParsers           error = errors.New("Missed parsers for queue processing")
var ErrorMissedConnection        error = errors.New("Missed rabbitmq connection")
var ErrorConnectionClosed        error = errors.New("Closed rabbitmq connection")
var ErrorConnectionRequired      error = errors.New("Create channel require connection")
var ErrorAlreadyConnected        error = errors.New("Connection for rabbitmq already established")

var ErrorUnprocessable           error = errors.New("Unprocessable entity")
var ErrorProcessingDuration      error = errors.New("Long entity processing")

var ErrorUnavailablePublisher    error = errors.New("Such publisher does not exist")

var ErrorMissedPublisherExchange error = errors.New("Publisher doesn't have exchange to push")

var ErrorLockForKeyNotFoundError error = errors.New("lock for key not found")
var ErrorLockForIdNotFoundError  error = errors.New("lock for id not found")
var ErrorLockMeetConflictError   error = errors.New("lock meet conflict")
