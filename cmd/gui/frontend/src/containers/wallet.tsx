import React, { useEffect, useState } from 'react';
import Frame from '../components/Frame';
import History from '../components/wallet/history';
import Receive from '../components/wallet/receive';
import Send from '../components/wallet/send';

type TabsProps = {
  tabs: string[];
  selectTab: (tab: string) => void;
  selectedTab?: string;
};

function Tabs({ tabs, selectTab, selectedTab }: TabsProps) {
  const isTabActive = (tab: string) => selectedTab === tab;
  return (
    <div className="wallet-tabs">
      {tabs.map((tab, idx) => (
        <div
          key={idx}
          className={`wallet-tab-${idx} ${
            isTabActive(tab) ? 'wallet-tab-active' : ''
          }`}
          onClick={() => selectTab(tab)}
        >
          {tab}
        </div>
      ))}
    </div>
  );
}

type TabBodyProps = {
  selectedTab: string;
  tabs: string[];
};

function TabBody({ selectedTab, tabs }: TabBodyProps) {
  if (selectedTab === tabs[0]) {
    return <History />;
  }

  if (selectedTab === tabs[1]) {
    return <Send />;
  }

  if (selectedTab === tabs[2]) {
    return <Receive />;
  }

  return <></>;
}

function Wallet() {
  const [selectedTab, setSelectedTab] = useState('');
  const [tabs] = useState(['Transaction History', 'Send', 'Receive']);

  const selectTab = (tab: string) => setSelectedTab(tab);

  useEffect(() => {
    selectTab(tabs[0]);
  }, [tabs]);

  return (
    <Frame header="wallet">
      <div id="wallet" className="page-container">
        <Tabs tabs={tabs} selectTab={selectTab} selectedTab={selectedTab} />
        <TabBody selectedTab={selectedTab} tabs={tabs} />
      </div>
    </Frame>
  );
}

export default Wallet;
