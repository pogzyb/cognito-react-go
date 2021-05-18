import './App.css';
import React from "react";
import { Route, Switch, BrowserRouter } from 'react-router-dom';
import jwt_decode from "jwt-decode";
import Home from "./Home";
import Token from "./Token";
import User from "./User";


function requireAuth(nextState, replace, next) {
    const isAuthed = this.getUser() !== null;
    const publicViews = ["/login", "/signup", "/forgot"];
    if ( !isAuthed && publicViews.includes(window.location.pathname) ) {
        console.log('here!');
        replace({
            pathname: "/login",
            state: {nextPathname: nextState.location.pathname}
        });
    }
    next();
}

class App extends React.Component {

    getUser() {
        const access_token = sessionStorage.getItem("access_token");
        console.log("access_token: ", access_token);
        if (access_token !== null) {
            let user = jwt_decode(access_token);
            console.log("user: ", user);
            return user
        }
        return null
    }

    render() {
        const user = this.getUser();
        const isAuthed = user !== null;
        return (
            <div className="App mt-5">
                <BrowserRouter>
                    <Switch>
                        <Route exact path='/' render={(props) => (
                            <Home {...props} isAuthed={isAuthed} />
                        )}/>
                        <Route exact path='/user' onenter={requireAuth}
                               render={(props) => (
                            <User {...props} username={user}/>
                        )}/>
                        <Route exact path='/authorize' component={Token}/>
                    </Switch>
                </BrowserRouter>
            </div>
        );
    }
}

export default App;
