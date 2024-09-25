import React from 'react';
import clsx from 'clsx';
import MDXContent from '@theme/MDXContent';
import ArchitectureContentMD from "./ArchitectureSection.content.md";
import styles from './ArchitectureSection.module.css';

export default function ArchitectureSectionView(): React.JSX.Element {
    return (
        <section className={clsx(styles.section)}>
            <div className={clsx("container padding-bottom--xl", styles.sectionDiv)}>
                <h2 id="architecture" className={clsx("text--center section__ref")}>Architecture</h2>
                <div className={clsx(styles.markdownContainer)}>
                    <MDXContent>
                        <ArchitectureContentMD/>
                    </MDXContent>
                </div>
            </div>
        </section>
    );
}
