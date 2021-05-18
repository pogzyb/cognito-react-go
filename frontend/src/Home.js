import React from "react";

const $ = window.$;

function getOpenStuff() {
    let url = window.location.origin + "/api-stuff/wide-open";
    $.ajax(url, {
        method: "GET",
        dataType: "json",
        success: (resp) => {
            $('#payload-display').html('<pre>' + resp["message"] + '</pre>');
        },
        error: (err) => {
            $('#payload-display').html(<pre> Oops! Bad response from the backend... </pre>);
            console.log(err);
        }
    });
}

class Home extends React.Component {

    constructor(props) {
        super(props);
        this.state = {isAuthed: props.isAuthed}
    }

    componentDidMount() {
        getOpenStuff();
    }

    render() {
        let btn = <a className="btn btn-primary" href="/">Sign-in <i className="fas fa-sign-in-alt"/></a>;
        if (this.state.isAuthed) {
            btn = <a className="btn btn-light" href="/user">View Account <i className="fas fa-user-ninja"/></a>;
        }
        return (
            <div className="container p-2">
                <h1 className="display-1">Welcome!</h1>
                <h1>This is the Homepage ...</h1>
                <p className="lead">
                    ... for a cool app that uses <i className="me-1 ms-1 fab fa-aws"/> Cognito for Authentication!
                </p>
                <div id="payload-display" className="mt-2 p-2">
                    <div className="spinner-grow" role="status">
                        <span className="sr-only">Loading...</span>
                    </div>
                </div>
                {btn}
            </div>
        );
    }
}

export default Home;