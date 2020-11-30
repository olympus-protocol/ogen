import React from 'react';
import AppFooter from './AppFooter';
import AppHeader from './AppHeader';
import AppSidebar from './AppSidebar';

interface Props {
    body: object,
    header: string
}

const AppFrame: React.FC<Props> = ({ body, header }) => {
    return (
        <div id="wrapper">
            <AppSidebar selected={header}/>
            <div id="wrapper-content">
                <AppHeader header={header}/>
                {body}
                <AppFooter />
            </div>
        </div>
    );
}

export default AppFrame;