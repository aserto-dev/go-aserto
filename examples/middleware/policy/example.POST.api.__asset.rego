package example.POST.api.__asset

import future.keywords.in

default allowed = false

allowed if {
    input.user.attributes.roles[_] == "editor"
    input.resource.asset != "secret"
}

allowed if {
    input.user.attributes.roles[_] == "admin"
}
