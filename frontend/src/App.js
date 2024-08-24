import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Navbar from './components/NavBar';
import ReportList from './components/ReportList';
import ReportDetail from './components/ReportDetail';

const App = () => {
  return (
    <Router>
      <Navbar />
      <Routes>
        <Route path="/" element={<ReportList />} />
        <Route path="/report/:reportId" element={<ReportDetail />} />
      </Routes>
    </Router>
  );
};
export default App;
