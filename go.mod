module go.seankhliao.com/webstyle

go 1.20

require github.com/yuin/goldmark v1.5.4

retract (
    [v0.4.0, v0.6.0] // go back to unversioned
    [v0.1.0, v0.4.0] // old version
)
