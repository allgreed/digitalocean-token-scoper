package main


type authorizationRequest struct {
    method string
    path string
}

type permission_rule interface {
    can_i() bool
    is_applicable() bool
}


type ble struct {}
