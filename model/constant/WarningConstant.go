package constant

const (
	LOG_ERROR_VALIDATE_DATA    = "Validasi Data"
	LOG_ERROR_VALIDATE_LDAP    = "Validasi LDAP"
	LOG_ERROR_CREATE_NEW_TOKEN = "Create New Token"
	LOG_ERROR_TOKEN_EXPIRED    = "Token Expired"
	LOG_ERROR_VALIDATE_DECRYPT = "Validasi Decrypt Data"

	LOG_ERROR_DESC_TOKEN_EXPIRED  = "token has been expired"
	LOG_ERROR_DESC_USER_NOT_FOUND = "user not found"

	ERR_APP_EXECPTION_TITLE         = "Problem with application"
	ERR_DATA_ACCESS_EXECPTION_TITLE = "Problem with database operation"
	ERR_OBJECT_VALIDATION_TITLE     = "Problem with data validation"
	ERR_OBJECT_VALIDATION_DETAIL    = "Input validation failed"
	ERR_MESSAGE_NOT_READABLE_TITLE  = "Problem with message not readable or in wrong format"
	ERR_MESSAGE_WRONG_FORMAT        = "Problem with file upload wrong format"

	ERR_MESSAGE_JSON_FORMAT  = "Json format you entered is wrong"
	ERR_MESSAGE_BINDING_DATA = "Please recheck your data binding in the database"
	ERR_MESSAGE_HEADER_TOKEN = "Header token cannot be empty"
)
