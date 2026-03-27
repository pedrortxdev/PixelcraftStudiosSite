import React from 'react';
import { NavLink } from 'react-router-dom';
import { LayoutDashboard, Wallet, Library, HeadphonesIcon, User } from 'lucide-react';
import './BottomNavigation.css';

const BottomNavigation = () => {
    return (
        <div className="bottom-navigation-container">
            <NavLink to="/dashboard" end className={({ isActive }) => `bottom-nav-item ${isActive ? 'active' : ''}`}>
                <LayoutDashboard size={24} />
                <span>Início</span>
            </NavLink>
            <NavLink to="/carteira" className={({ isActive }) => `bottom-nav-item ${isActive ? 'active' : ''}`}>
                <Wallet size={24} />
                <span>Carteira</span>
            </NavLink>
            <NavLink to="/projetos" className={({ isActive }) => `bottom-nav-item ${isActive ? 'active' : ''}`}>
                <Library size={24} />
                <span>Projetos</span>
            </NavLink>
            <NavLink to="/suporte" className={({ isActive }) => `bottom-nav-item ${isActive ? 'active' : ''}`}>
                <HeadphonesIcon size={24} />
                <span>Suporte</span>
            </NavLink>
            <NavLink to="/configuracoes" className={({ isActive }) => `bottom-nav-item ${isActive ? 'active' : ''}`}>
                <User size={24} />
                <span>Perfil</span>
            </NavLink>
        </div>
    );
};

export default BottomNavigation;
