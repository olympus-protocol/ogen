import React, { ReactNode } from 'react';
import Footer from '../Footer';
import Header from '../Header';
import Sidebar from '../Sidebar';

interface FrameProps {
  header: string;
  children?: ReactNode;
}

export default function Frame({ header, children }: FrameProps) {
  return (
    <div id="wrapper">
      <Sidebar selected={header} />
      <div id="wrapper-content">
        <Header header={header} />
        {children}
        <Footer />
      </div>
    </div>
  );
}
