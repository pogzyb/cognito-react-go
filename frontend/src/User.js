import React from "react";

const $ = window.$;

function getSecureStuff() {
    let url = window.location.origin + "/api-stuff/top-secret";
    $.ajax(url, {
        method: "GET",
        dataType: "json",
        headers: {
            "Authorization": "Bearer " + sessionStorage.getItem("access_token")
        },
        success: (resp) => {
            $('#payload-display').html('<pre>' + resp["message"] + '</pre>');
        },
        error: (err) => {
            $('#payload-display').html(<pre> Oops! Bad response from the backend... </pre>);
            console.log(err);
        }
    });
}

function getUserInfo() {
    let url = window.location.origin + "/api-stuff/user-info";
    $.ajax(url, {
        method: "GET",
        dataType: "json",
        headers: {
            "Authorization": "Bearer " + sessionStorage.getItem("access_token")
        },
        success: (resp) => {
            $('#user-display').html('<pre>' + JSON.stringify(resp["message"], null, 2) + '</pre>');
        },
        error: (err) => {
            $('#user-display').html(<pre> Oops! Bad response from the backend... </pre>);
            console.log(err);
        }
    });
}

class User extends React.Component {

    componentDidMount() {
        getSecureStuff();
        getUserInfo();
    }

    render() {
        return (
            <div className="container">
                <p className="lead">Secret users only message:</p>
                <div className="mt-2 p-2" id="payload-display">
                    <div className="spinner-grow" role="status">
                        <span className="sr-only">Loading...</span>
                    </div>
                </div>
                <p className="lead">UserInfo:</p>
                <div className="mt-2 p-2" id="user-display">
                    <div className="spinner-grow" role="status">
                        <span className="sr-only">Loading...</span>
                    </div>
                </div>
                <a className="btn btn-light" href="/">Home <i className="fas fa-home"/></a>
            </div>
        );
    }
}

export default User;