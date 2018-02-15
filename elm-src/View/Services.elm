module View.Services exposing (view)

import Model exposing (Model)
import Types exposing (Service)
import Msg exposing (Msg(..))
import Html exposing (Html, text)
import Material.List as List


view : Model -> Html Msg
view model =
    List.ul []
        (List.map (viewServiceRow model) model.services)


viewServiceRow : Model -> Service -> Html Msg
viewServiceRow model service =
    List.li []
        [ List.content []
            [ text service.name ]
        ]
