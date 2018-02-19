module Route exposing (..)

import Navigation
import UrlParser exposing (parsePath, oneOf, map, top, s, (</>), string, parseHash)


type Route
    = Home
    | Users
    | Services


type alias Model =
    Maybe Route


pathParser : UrlParser.Parser (Route -> a) a
pathParser =
    oneOf
        [ map Home top
        , map Users (s "users")
        , map Services (s "services")
        ]


init : Maybe Route -> List (Maybe Route)
init location =
    case location of
        Nothing ->
            [ Just Services ]

        something ->
            [ something ]


urlFor : Route -> String
urlFor loc =
    case loc of
        Home ->
            "#"

        Users ->
            "#users"

        Services ->
            "#services"


locFor : Navigation.Location -> Maybe Route
locFor path =
    parseHash pathParser path
