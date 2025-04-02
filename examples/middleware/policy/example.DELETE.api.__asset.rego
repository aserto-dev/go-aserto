package example.DELETE.api.__asset

import future.keywords.in

default allowed = false

allowed if {
    input.user.attributes.roles[_] == "admin"
}
