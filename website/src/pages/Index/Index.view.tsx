import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HeroSection from '@site/src/components/Index/HeroSection';
import WhatIsItSection from '@site/src/components/Index/WhatIsItSection';
import ArchitectureSection from '@site/src/components/Index/ArchitectureSection';
import CommunitySection from '@site/src/components/Index/CommunitySection';


export default function IndexView(): React.JSX.Element {
    const {siteConfig} = useDocusaurusContext();
    return (
        <Layout
            title="Welcome"
            description={siteConfig.tagline}>
            <HeroSection/>
            <main>
                <WhatIsItSection/>
                <ArchitectureSection/>
                <CommunitySection/>
            </main>
        </Layout>
    );
}
