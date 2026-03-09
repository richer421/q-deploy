import {
  ApiOutlined,
  CloudServerOutlined,
  CodeOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  DeploymentUnitOutlined,
  RocketOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { Col, Row, Tag } from 'antd';
import React from 'react';
import styles from './index.less';

const LAYERS = [
  { key: 'http', label: 'HTTP', desc: '接口层', detail: 'Gin + Swagger', color: '#1677FF' },
  { key: 'app', label: 'App', desc: '应用层', detail: '用例编排', color: '#722ED1' },
  { key: 'domain', label: 'Domain', desc: '领域层', detail: '核心业务', color: '#FA8C16' },
  { key: 'infra', label: 'Infra', desc: '基础设施', detail: 'MySQL / Redis / Kafka', color: '#52C41A' },
];

const TECH = [
  { label: 'Go', color: '#00ADD8' },
  { label: 'Gin', color: '#00B386' },
  { label: 'GORM Gen', color: '#E44D26' },
  { label: 'Cobra', color: '#ED6A5A' },
  { label: 'MySQL', color: '#4479A1' },
  { label: 'Redis', color: '#DC382D' },
  { label: 'Kafka', color: '#231F20' },
  { label: 'OpenTelemetry', color: '#F5A800' },
  { label: 'Ant Design', color: '#1677FF' },
  { label: 'UmiJS', color: '#1890FF' },
];

const CAPS = [
  { icon: <ApiOutlined />, title: 'RESTful API', desc: '版本化路由 · Swagger 文档' },
  { icon: <DeploymentUnitOutlined />, title: 'DDD 分层', desc: '严格单向依赖 · 清晰边界' },
  { icon: <DatabaseOutlined />, title: '数据基础设施', desc: 'MySQL + Redis + Kafka' },
  { icon: <DashboardOutlined />, title: '可观测性', desc: 'OpenTelemetry · Prometheus' },
  { icon: <ThunderboltOutlined />, title: '代码生成', desc: 'GORM Gen 类型安全查询' },
  { icon: <CloudServerOutlined />, title: '容器部署', desc: 'Docker Compose 一键编排' },
  { icon: <CodeOutlined />, title: 'AI 协作', desc: 'Knowledge 文档驱动' },
  { icon: <RocketOutlined />, title: '工程化', desc: 'Makefile · golangci-lint' },
];

const Home: React.FC = () => {
  return (
    <div className={styles.page}>
      {/* Hero */}
      <div className={styles.hero}>
        <div className={styles.heroMain}>
          <div className={styles.brand}>
            <span className={styles.brandQ}>Q</span>
            <span className={styles.brandDash}>-</span>
            <span className={styles.brandDev}>DEV</span>
          </div>
          <div className={styles.tagline}>AI 驱动的全栈开发脚手架</div>
          <div className={styles.subtitle}>
            让 AI 理解你的架构，让开发回归本质
          </div>
          <div className={styles.tags}>
            <Tag color="blue">v0.1.0</Tag>
            <span className={styles.techHint}>Go · Gin · GORM · Ant Design Pro</span>
          </div>
        </div>
        <div className={styles.heroSide}>
          <div className={styles.codeBlock}>
            <div className={styles.codeLine}>
              <span className={styles.codeComment}># 启动全栈服务</span>
            </div>
            <div className={styles.codeLine}>
              <span className={styles.codeCmd}>make docker-up</span>
            </div>
            <div className={styles.codeBlank} />
            <div className={styles.codeLine}>
              <span className={styles.codeComment}># 本地开发</span>
            </div>
            <div className={styles.codeLine}>
              <span className={styles.codeCmd}>make dev</span>
              <span className={styles.codeComment}># 后端</span>
            </div>
            <div className={styles.codeLine}>
              <span className={styles.codeCmd}>make fe-dev</span>
              <span className={styles.codeComment}># 前端</span>
            </div>
          </div>
        </div>
      </div>

      {/* Architecture */}
      <div className={styles.section}>
        <div className={styles.sectionTitle}>
          <span className={styles.sectionIcon}>◈</span>
          DDD 分层架构
        </div>
        <div className={styles.archFlow}>
          {LAYERS.map((layer, i) => (
            <React.Fragment key={layer.key}>
              <div className={styles.archNode}>
                <div className={styles.archNodeLine} style={{ background: layer.color }} />
                <div className={styles.archNodeLabel} style={{ color: layer.color }}>
                  {layer.label}
                </div>
                <div className={styles.archNodeDesc}>{layer.desc}</div>
                <div className={styles.archNodeDetail}>{layer.detail}</div>
              </div>
              {i < LAYERS.length - 1 && <div className={styles.archArrow}>→</div>}
            </React.Fragment>
          ))}
        </div>
        <div className={styles.archNote}>
          <CodeOutlined /> 严格单向依赖：http → app → domain → infra
        </div>
      </div>

      {/* Tech Stack */}
      <div className={styles.section}>
        <div className={styles.sectionTitle}>
          <span className={styles.sectionIcon}>◈</span>
          技术栈
        </div>
        <div className={styles.techGrid}>
          {TECH.map((t) => (
            <div key={t.label} className={styles.techBadge}>
              <span className={styles.techDot} style={{ background: t.color }} />
              {t.label}
            </div>
          ))}
        </div>
      </div>

      {/* Capabilities */}
      <div className={styles.section}>
        <div className={styles.sectionTitle}>
          <span className={styles.sectionIcon}>◈</span>
          核心能力
        </div>
        <Row gutter={[16, 16]}>
          {CAPS.map((cap) => (
            <Col key={cap.title} xs={24} sm={12} lg={6}>
              <div className={styles.capCard}>
                <div className={styles.capIcon}>{cap.icon}</div>
                <div className={styles.capTitle}>{cap.title}</div>
                <div className={styles.capDesc}>{cap.desc}</div>
              </div>
            </Col>
          ))}
        </Row>
      </div>
    </div>
  );
};

export default Home;
