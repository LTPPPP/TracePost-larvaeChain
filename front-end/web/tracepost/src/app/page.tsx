import styles from './home.module.scss';
import classNames from 'classnames/bind';
const cx = classNames.bind(styles);

function Home() {
  return <div className={cx('title', 'bg-sky-950', 'p-8')}>hello world</div>;
}
export default Home;
