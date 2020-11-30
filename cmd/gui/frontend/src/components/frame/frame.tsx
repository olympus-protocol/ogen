import React from 'react';
import Footer from '../footer/footer';
import Header from '../header/header';
import Sidebar from '../sidebar/sidebar';

interface Props {
    body: object,
    header: string
}

const Frame: React.FC<Props> = ({ body, header }) => {
    return (
        <div id="wrapper">
            <Sidebar selected={header}/>
            <div id="wrapper-content">
                <Header header={header}/>
                {body}
                <Footer />
            </div>
        </div>
    );
}

export default Frame;