import React from 'react';
import clsx from 'clsx';
import styles from './CommunitySection.module.css';
import {IconMessages} from "@tabler/icons-react";
import Link from "@docusaurus/Link";


export default function CommunitySectionView(): React.JSX.Element {
    return (
        <section>
            <div className={clsx("container padding-bottom--xl padding-top--lg text--center", styles.sectionDiv)}>
                <h2 id="community" className={clsx("section__ref")}>Community</h2>
                <p className={clsx("padding-horiz--xl", styles.description)}>
                    The Yunikorn History Server is an open-source project and we welcome contributions from the community. If you have any
                    questions, suggestions, or feedback, please feel free to react out to us on GitHub.
                </p>
                <Link
                    className="button button--primary button--lg"
                    to="https://github.com/G-Research/yunikorn-history-server/discussions/new/choose">
                    <IconMessages size={18} color="currentcolor" stroke={3}/> Start a Discussion
                </Link>
            </div>
        </section>
    );
}
