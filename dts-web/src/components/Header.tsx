import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Flex, Text, Button } from '@radix-ui/themes';

const Header: React.FC = () => {
    const location = useLocation();

    return (
        <Flex asChild justify="between" align="center" p="4" style={{ borderBottom: '1px solid #eaeaea' }}>
            <header>
                <Text size="5" weight="bold">Distributed Task Scheduler</Text>
                <Flex gap="4">
                    <Button asChild variant={location.pathname === '/' ? 'solid' : 'soft'}>
                        <Link to="/">Job List</Link>
                    </Button>
                    <Button asChild variant={location.pathname === '/create' ? 'solid' : 'soft'}>
                        <Link to="/create">Create Job</Link>
                    </Button>
                </Flex>
            </header>
        </Flex>
    );
};

export default Header;
