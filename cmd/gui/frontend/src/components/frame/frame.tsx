import React from 'react';
import { Footer } from '../footer/footer';
import Header from '../header/header';
import { Sidebar } from '../sidebar/sidebar';

interface FrameProps {
    body: object,
    header: string
}

export default class Frame extends React.Component<FrameProps, any> {
    render() {
        return (
            <div id="wrapper">
                <Sidebar selected={this.props.header}/>
                <div id="wrapper-content">
                    <Header header={this.props.header}/>
                    {this.props.body}
                    <Footer />
                </div>
            </div>
        );
    }
}
