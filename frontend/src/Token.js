import React from "react";

const $ = window.$;

function getAndSetTokens(code) {
    let url = window.location.origin + "/api-stuff/authorize" + code;
    $.ajax(url, {
        method: "GET",
        dataType: "json",
        success: (resp) => {
            sessionStorage.setItem("access_token", resp["access_token"]);
            window.location.replace("/");
        },
        error: (err) => {
            console.log(err);
        }
    });
}

class Token extends React.Component {

    componentDidMount() {
        getAndSetTokens(this.props.location.search);
    }

    render() {
        return (
            <div className="container">
                <div className="mt-5 p-5">
                    <h2>Loading your space...</h2>
                    <div className="spinner-grow" role="status">
                        <span className="sr-only">Loading...</span>
                    </div>
                </div>
            </div>
        );
    }
}

export default Token;
