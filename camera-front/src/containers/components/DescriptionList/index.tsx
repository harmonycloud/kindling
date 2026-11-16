import React, { Component } from 'react';
import styles from "./index.module.less";

// 占用的空间大小
type Space = "quater" | "half" | "full";

export interface Specification {
  name: string;
  key: string;
  space?: Space;
  render?: (value: any, allObj: any) => React.ReactNode;
}

interface Props {
  title?: string;
  data: any;
  style?: React.CSSProperties;
  specifications: Specification[];
}

export default class DescriptionList extends Component<Props, any> {
  constructor(props) {
    super(props);
    this.state = {
      flag: 1
    };
  }

  render() {
    const { specifications, data, title, style } = this.props;
    return (
      <>
        <div style={style} className={styles.container}>
          <span className={styles.title}>{title ? title : ""}</span>
          {
            specifications.length === 0 ? '无数据' : ''
          }
          {specifications.map(specification => specification ? <span
            key={specification.key}
            className={
              styles.item + " " +
              styles[specification.space ? specification.space : "half"]
            }
          >
            <span className={styles.key}>
              {specification.name}:
            </span>
            <span className={styles.value}>
              {
                typeof specification.render === "function" ?
                  specification.render(data[specification.key], data)
                  : (data[specification.key] ?? "--")
              }
            </span>
          </span> : "")}
        </div>
      </>
    );
  }
}

