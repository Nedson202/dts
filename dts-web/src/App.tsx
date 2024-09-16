import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import { Theme } from '@radix-ui/themes';
import '@radix-ui/themes/styles.css';
import Header from './components/Header';
import JobList from './components/JobList';
import JobCreation from './components/JobCreation';
import JobDetails from './components/JobDetails';
import { Spinner } from './components/Spinner';
import { useAuth } from '@workos-inc/authkit-react';

function App() {
    const { user, getAccessToken, isLoading, signIn, signUp, signOut } = useAuth();
    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-screen">
                <Spinner />
            </div>
        )
    }

    const performMutation = async () => {
        const accessToken = await getAccessToken();
        console.log("api request with accessToken", accessToken);
    };

    return (
        <Theme>
            <Router>
                <div className="App">
                    <Header />
                    <main style={{ padding: '20px' }}>
                        <Routes>
                            <Route path="/" element={<JobList />} />
                            <Route path="/create" element={<JobCreation />} />
                            <Route path="/job/:id" element={<JobDetails />} />
                        </Routes>
                    </main>
                </div>
            </Router>
        </Theme>
    );
}

export default App;
