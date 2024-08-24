import React from 'react';
import { Navbar as BootstrapNavbar, Container } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { AiOutlineRobot } from 'react-icons/ai';

const Navbar = () => {
  return (
    <BootstrapNavbar bg="dark" variant="dark">
      <Container>
        <Link to="/" className="navbar-brand">
          <AiOutlineRobot size={50} /> <span>Dexory Report</span>
        </Link>
      </Container>
    </BootstrapNavbar>
  );
};

export default Navbar;
