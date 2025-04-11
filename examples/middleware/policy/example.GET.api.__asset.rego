package example.GET.api.__asset

import future.keywords.in

default allowed = false

allowed if {
    roles := {"viewer", "editor"}

    some x in roles
    input.user.attributes.roles[_] == x

    input.resource.asset != "secret"
}

allowed if {
    input.user.attributes.roles[_] == "admin"
}
