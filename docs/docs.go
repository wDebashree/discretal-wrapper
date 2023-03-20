// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://iot.discretal.com/terms/",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/channels": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves a list of channels. Due to performance concerns, data is retrieved in subsets. The API things must ensure that the entire dataset is consumed either by making subsequent requests, or by increasing the subset size of the initial request.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "channels"
                ],
                "summary": "Retrieves channels",
                "parameters": [
                    {
                        "type": "integer",
                        "default": 100,
                        "description": "Size of the subset to retrieve.",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "default": 0,
                        "description": "Number of items to skip during retrieval.",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Data retrieved.",
                        "schema": {
                            "$ref": "#/definitions/models.ChannelsList"
                        }
                    },
                    "400": {
                        "description": "Failed due to malformed query parameters."
                    },
                    "401": {
                        "description": "Missing or invalid access token provided."
                    },
                    "500": {
                        "description": "Unexpected server-side error occurred."
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Creates new channel. User identified by the provided access token will be the channels owner.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "channels"
                ],
                "summary": "Adds new channel",
                "parameters": [
                    {
                        "description": "JSON-formatted document describing the updated channel.",
                        "name": "Request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.ChannelReq"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Channel created.",
                        "schema": {
                            "$ref": "#/definitions/models.ChannelRes"
                        }
                    },
                    "400": {
                        "description": "Failed due to malformed JSON."
                    },
                    "401": {
                        "description": "Missing or invalid access token provided."
                    },
                    "500": {
                        "description": "Unexpected server-side error occurred."
                    }
                }
            }
        },
        "/channels/{name}": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves the details of a channel",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "channels"
                ],
                "summary": "Retrieves channel info",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Unique channel name.",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Data retrieved.",
                        "schema": {
                            "$ref": "#/definitions/models.ChannelRes"
                        }
                    },
                    "400": {
                        "description": "Failed due to malformed channel's ID."
                    },
                    "401": {
                        "description": "Missing or invalid access token provided."
                    },
                    "404": {
                        "description": "Channel does not exist."
                    },
                    "500": {
                        "description": "Unexpected server-side error occurred."
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Removes a channel. The service will ensure that the subscribed apps and things are unsubscribed from the removed channel.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "channels"
                ],
                "summary": "Removes a channel",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Unique channel name.",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Channel removed."
                    },
                    "400": {
                        "description": "Failed due to malformed channel's ID."
                    },
                    "401": {
                        "description": "Missing or invalid access token provided."
                    },
                    "500": {
                        "description": "Unexpected server-side error occurred."
                    }
                }
            }
        },
        "/login": {
            "post": {
                "description": "Generates an access token when provided with proper credentials.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "User authentication",
                "parameters": [
                    {
                        "description": "JSON-formatted document describing the user details for login",
                        "name": "Request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.LoginUserReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "User authenticated.",
                        "schema": {
                            "$ref": "#/definitions/models.LoginUserRes"
                        }
                    },
                    "400": {
                        "description": "Failed due to malformed JSON."
                    },
                    "500": {
                        "description": "Unexpected server-side error occurred."
                    }
                }
            }
        },
        "/register": {
            "post": {
                "description": "Registers new user account given email and password. New account will be uniquely identified by its email address.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Registers user account",
                "parameters": [
                    {
                        "description": "JSON-formatted document describing the new user to be registered",
                        "name": "Request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.RegisterUserReq"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Registered new user.",
                        "schema": {
                            "$ref": "#/definitions/models.RegisterUserRes"
                        }
                    },
                    "400": {
                        "description": "Failed due to malformed JSON."
                    },
                    "500": {
                        "description": "Unexpected server-side error occurred."
                    }
                }
            }
        },
        "/things": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves a list of things. Due to performance concerns, data is retrieved in subsets. The API things must ensure that the entire dataset is consumed either by making subsequent requests, or by increasing the subset size of the initial request.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "things"
                ],
                "summary": "Retrieves things",
                "parameters": [
                    {
                        "type": "integer",
                        "default": 100,
                        "description": "Size of the subset to retrieve.",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "default": 0,
                        "description": "Number of items to skip during retrieval.",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Data retrieved.",
                        "schema": {
                            "$ref": "#/definitions/models.ThingsList"
                        }
                    },
                    "400": {
                        "description": "Failed due to malformed query parameters."
                    },
                    "401": {
                        "description": "Missing or invalid access token provided."
                    },
                    "500": {
                        "description": "Unexpected server-side error occurred."
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Adds new thing to the list of things owned by user identified using the provided access token.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "things"
                ],
                "summary": "Adds new thing",
                "parameters": [
                    {
                        "description": "JSON-formatted document describing the new thing.",
                        "name": "Request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.ThingReq"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Thing registered.",
                        "schema": {
                            "$ref": "#/definitions/models.ThingRes"
                        }
                    },
                    "400": {
                        "description": "Failed due to malformed JSON."
                    },
                    "401": {
                        "description": "Missing or invalid access token provided."
                    },
                    "500": {
                        "description": "Unexpected server-side error occurred."
                    }
                }
            }
        },
        "/things/{name}": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves the details of a thing",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "things"
                ],
                "summary": "Retrieves thing info",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Unique thing name.",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Data retrieved.",
                        "schema": {
                            "$ref": "#/definitions/models.ThingRes"
                        }
                    },
                    "400": {
                        "description": "Failed due to malformed thing's ID."
                    },
                    "401": {
                        "description": "Missing or invalid access token provided."
                    },
                    "404": {
                        "description": "Thing does not exist."
                    },
                    "500": {
                        "description": "Unexpected server-side error occurred."
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Removes a thing. The service will ensure that the removed thing is disconnected from all of the existing channels.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "things"
                ],
                "summary": "Removes a thing",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Unique thing name.",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Thing removed."
                    },
                    "400": {
                        "description": "Failed due to malformed thing's ID."
                    },
                    "401": {
                        "description": "Missing or invalid access token provided."
                    },
                    "500": {
                        "description": "Unexpected server-side error occurred."
                    }
                }
            }
        }
    },
    "definitions": {
        "models.ChannelReq": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "metadata": {
                    "$ref": "#/definitions/models.Metadata"
                },
                "name": {
                    "type": "string",
                    "example": "channel1"
                }
            }
        },
        "models.ChannelRes": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string",
                    "example": "880d7429-8857-4e50-a7e0-698e2865b0aa"
                },
                "metadata": {
                    "$ref": "#/definitions/models.Metadata"
                },
                "name": {
                    "type": "string",
                    "example": "channel1"
                }
            }
        },
        "models.ChannelsList": {
            "type": "object",
            "properties": {
                "channels": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.ChannelRes"
                    }
                }
            }
        },
        "models.LoginUserReq": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string",
                    "example": "user1@example.com"
                },
                "password": {
                    "type": "string",
                    "example": "pass@1234"
                }
            }
        },
        "models.LoginUserRes": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string",
                    "example": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkaXNjcmV0YWwuYXV0aCIsInN1YiI6InVzZXIxQGV4YW1wbGUuY29tIiwiZXhwIjoxNjcyMDkyNDYzLCJpYXQiOjE2NzIwNTY0NjMsImlzc3Vlcl9pZCI6ImY5ZGJiZjIyLTcxZWQtNGIxZC1hZTU3LTk3ZjIxYjA4YTJiOSIsInR5cGUiOjB9.-Lcm4eWaR82W_oEVIgB24-ao6kI2NE80qR-nAiwh_c8"
                }
            }
        },
        "models.Metadata": {
            "type": "object",
            "additionalProperties": true
        },
        "models.RegisterUserReq": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string",
                    "example": "user1@example.com"
                },
                "metadata": {
                    "$ref": "#/definitions/models.Metadata"
                },
                "password": {
                    "type": "string",
                    "example": "pass@1234"
                }
            }
        },
        "models.RegisterUserRes": {
            "type": "object",
            "required": [
                "id"
            ],
            "properties": {
                "id": {
                    "type": "string"
                }
            }
        },
        "models.ThingReq": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "metadata": {
                    "$ref": "#/definitions/models.Metadata"
                },
                "name": {
                    "type": "string",
                    "example": "device1"
                }
            }
        },
        "models.ThingRes": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string",
                    "example": "8c0c7129-8857-4e50-a7e0-698e2865b0aa"
                },
                "key": {
                    "type": "string",
                    "example": "ef751d71-fb43-423c-a2eb-8602e6232cb4"
                },
                "metadata": {
                    "$ref": "#/definitions/models.Metadata"
                },
                "name": {
                    "type": "string",
                    "example": "device1"
                }
            }
        },
        "models.ThingsList": {
            "type": "object",
            "properties": {
                "things": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.ThingRes"
                    }
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:5000",
	BasePath:         "/api",
	Schemes:          []string{"http"},
	Title:            "Discretal API",
	Description:      "A wrapper api for utilizing Discretal server messaging services over MQTT",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}